package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/services"
	"github.com/Clinet/discordgo-embed"
	"github.com/JoshuaDoes/logger"
)

var Log *logger.Logger
var Discord *ClientDiscord

//ClientDiscord implements services.Service and holds a Discord session
type ClientDiscord struct {
	*discordgo.Session
	User *discordgo.User
	VCs  []*discordgo.VoiceConnection
}

func (discord *ClientDiscord) Shutdown() {
	for _, vc := range discord.VCs {
		_ = vc.Disconnect()
	}
	_ = discord.Close()
}

func (discord *ClientDiscord) CmdPrefix() string {
	return "/"
}

func (discord *ClientDiscord) Login(cfg interface{}) (err error) {
	DiscordCfg = cfg.(*CfgDiscord)
	Log.Trace("--- StartDiscord() ---")

	Log.Debug("Creating Discord struct...")
	discordClient, err := discordgo.New("Bot " + DiscordCfg.Token)
	if err != nil {
		return err
	}

	//Only enable informational Discord logging if we're tracing
	if Log.Verbosity == 2 {
		Log.Debug("Setting Discord log level to informational...")
		discordClient.LogLevel = discordgo.LogInformational
	}

	Log.Info("Registering Discord event handlers...")
	discordClient.AddHandler(discordReady)
	discordClient.AddHandler(discordMessageCreate)
	discordClient.AddHandler(discordInteractionCreate)

	Log.Info("Connecting to Discord...")
	err = discordClient.Open()
	if err != nil {
		return err
	}

	Log.Info("Connected to Discord!")
	Discord = &ClientDiscord{discordClient, nil, make([]*discordgo.VoiceConnection, 0)}

	Log.Info("Recycling old application commands...")
	if oldAppCmds, err := Discord.ApplicationCommands(Discord.State.User.ID, ""); err == nil {
		for _, cmd := range oldAppCmds {
			Log.Trace("Deleting application command for ", cmd.Name)
			if err := Discord.ApplicationCommandDelete(Discord.State.User.ID, "", cmd.ID); err != nil {
				return err
			}
		}
	}

	Log.Info("Registering application commands...")
	Log.Warn("TODO: Batch overwrite commands, then get a list of commands from Discord that aren't in memory and delete them")
	for _, cmd := range CmdsToAppCommands() {
		Log.Trace("Registering cmd: ", cmd)
		_, err := Discord.ApplicationCommandCreate(Discord.State.User.ID, "", cmd)
		if err != nil {
			Log.Fatal(services.Error("Unable to register cmd '%s': %v", cmd.Name, err))
		}
	}
	Log.Info("Application commands ready for use!")
	return nil
}

func (discord *ClientDiscord) MsgEdit(msg *services.Message) (ret *services.Message, err error) {
	return nil, nil
}
func (discord *ClientDiscord) MsgRemove(msg *services.Message) (err error) {
	return nil
}
func (discord *ClientDiscord) MsgSend(msg *services.Message) (ret *services.Message, err error) {
	msgContext := msg.Context
	switch msgContext.(type) {
	case *discordgo.Message:
		if msg.ChannelID != "" {
			return nil, services.Error("discord: MsgSend(msg: %v): missing channel ID", msg)
		}
	case *discordgo.Interaction:
		if msg.MessageID != "" {
			return nil, services.Error("discord: MsgSend(msg: %v): missing interaction ID as message ID", msg)
		}
	default:
		//Sending a DM to a user should always be a regular message
		if msg.ServerID == "" && msg.ChannelID != "" {
			channelDM, err := discord.UserChannelCreate(msg.ChannelID)
			if err != nil {
				return nil, services.Error("discord: MsgSend(msg: %v): unable to create DM with userID: %s: %v", msg, msg.ChannelID, err)
			}
			msg.ChannelID = channelDM.ID
		} else {
			return nil, services.Error("discord: MsgSend(msg: %v): unknown MsgContext: %d", msg, msgContext)
		}
	}

	var discordMsg *discordgo.Message
	if msg.Title != "" || msg.Color != nil || msg.Image != "" {
		retEmbed := embed.NewEmbed().SetDescription(msg.Content)
		if msg.Title != "" {
			retEmbed.SetTitle(msg.Title)
		}
		if msg.Color != nil {
			retEmbed.SetColor(*msg.Color)
		}
		if msg.Image != "" {
			retEmbed.SetImage(msg.Image)
		}

		switch msgContext.(type) {
		case *discordgo.Interaction:
			interactionResp := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{retEmbed.MessageEmbed},
				},
			}
			interaction := msg.Context.(*discordgo.Interaction)
			err = discord.InteractionRespond(interaction, interactionResp)
		default:
			discordMsg, err = discord.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{Embed: retEmbed.MessageEmbed})
		}
	} else {
		if msg.Content == "" {
			return nil, services.Error("discord: MsgSend(msg: %v): missing content", msg)
		}

		switch msgContext.(type) {
		case *discordgo.Interaction:
			interactionResp := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg.Content,
				},
			}
			interaction := msg.Context.(*discordgo.Interaction)
			err = discord.InteractionRespond(interaction, interactionResp)
		default:
			discordMsg, err = discord.ChannelMessageSend(msg.ChannelID, msg.Content)
		}
	}
	if err != nil {
		return nil, err
	}

	ret = msg
	if discordMsg != nil {
		ret.UserID = discordMsg.Author.ID
		ret.ServerID = discordMsg.GuildID
	}
	if discordMsg != nil {
		ret = &services.Message{
			UserID: discordMsg.Author.ID,
			MessageID: discordMsg.ID,
			ChannelID: discordMsg.ChannelID,
			ServerID: discordMsg.GuildID,
			Content: discordMsg.Content,
			Context: discordMsg,
		}
	}
	return ret, err
}

func (discord *ClientDiscord) GetUser(serverID, userID string) (ret *services.User, err error) {
	user, err := discord.GuildMember(serverID, userID)
	if err != nil {
		return nil, err
	}

	userRoles := make([]*services.Role, len(user.Roles))
	for i := 0; i < len(userRoles); i++ {
		role := &services.Role{
			RoleID: user.Roles[i],
		}
		userRoles[i] = role
	}
	return &services.User{
		ServerID: serverID,
		UserID: userID,
		Username: user.User.Username,
		Nickname: user.Nick,
		Roles: userRoles,
	}, nil
}
func (discord *ClientDiscord) GetUserPerms(serverID, channelID, userID string) (perms *services.Perms, err error) {
	user, err := discord.GetUser(serverID, userID)
	if err != nil {
		return nil, err
	}

	server, err := discord.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	guildRoles, err := discord.GuildRoles(serverID)
	if err != nil {
		return nil, err
	}

	channel, err := discord.Channel(channelID)
	if err != nil {
		return nil, err
	}

	perms = &services.Perms{}
	for i := 0; i < len(guildRoles); i++ {
		for j := 0; j < len(user.Roles); j++ {
			if guildRoles[i].ID == user.Roles[j].RoleID {
				guildRolePerms := guildRoles[i].Permissions
				if guildRolePerms & discordgo.PermissionAdministrator != 0 {
					perms.Administrator = true
				}
				if guildRolePerms & discordgo.PermissionKickMembers != 0 {
					perms.Kick = true
				}
				if guildRolePerms & discordgo.PermissionBanMembers != 0 {
					perms.Ban = true
				}

				for _, overwrite := range channel.PermissionOverwrites {
					if overwrite.Type == discordgo.PermissionOverwriteTypeRole && overwrite.ID == guildRoles[i].ID {
						if overwrite.Allow & discordgo.PermissionAdministrator != 0 {
							perms.Administrator = true
						}
						if overwrite.Allow & discordgo.PermissionKickMembers != 0 {
							perms.Kick = true
						}
						if overwrite.Allow & discordgo.PermissionBanMembers != 0 {
							perms.Ban = true
						}
						if overwrite.Deny & discordgo.PermissionAdministrator != 0 {
							perms.Administrator = false
						}
						if overwrite.Deny & discordgo.PermissionKickMembers != 0 {
							perms.Kick = false
						}
						if overwrite.Deny & discordgo.PermissionBanMembers != 0 {
							perms.Ban = false
						}
					}
				}
			}
		}
	}

	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == discordgo.PermissionOverwriteTypeMember && overwrite.ID == userID {
			if overwrite.Allow & discordgo.PermissionAdministrator != 0 {
				perms.Administrator = true
			}
			if overwrite.Allow & discordgo.PermissionKickMembers != 0 {
				perms.Kick = true
			}
			if overwrite.Allow & discordgo.PermissionBanMembers != 0 {
				perms.Ban = true
			}
			if overwrite.Deny & discordgo.PermissionAdministrator != 0 {
				perms.Administrator = false
			}
			if overwrite.Deny & discordgo.PermissionKickMembers != 0 {
				perms.Kick = false
			}
			if overwrite.Deny & discordgo.PermissionBanMembers != 0 {
				perms.Ban = false
			}
		}
	}

	if server.OwnerID == userID {
		perms.Administrator = true
	}

	return perms, nil
}
func (discord *ClientDiscord) UserBan(user *services.User, reason string, rule int) (err error) {
	Log.Trace("Ban(", user.ServerID, ", ", user.UserID, ", ", reason, ", ", rule, ")")
	return discord.GuildBanCreateWithReason(user.ServerID, user.UserID, reason, 0)
}
func (discord *ClientDiscord) UserKick(user *services.User, reason string, rule int) (err error) {
	Log.Trace("Kick(", user.ServerID, ", ", user.UserID, ", ", reason, ", ", rule, ")")
	return discord.GuildMemberDeleteWithReason(user.ServerID, user.UserID, reason)
}

func (discord *ClientDiscord) GetServer(serverID string) (server *services.Server, err error) {
	guild, err := discord.State.Guild(serverID)
	if err != nil {
		return nil, err
	}

	voiceStates := make([]*services.VoiceState, len(guild.VoiceStates))
	for i := 0; i < len(voiceStates); i++ {
		vs := guild.VoiceStates[i]
		voiceStates[i] = &services.VoiceState{
			ChannelID: vs.ChannelID,
			UserID: vs.UserID,
			SessionID: vs.SessionID,
			Deaf: vs.Deaf,
			Mute: vs.Mute,
			SelfDeaf: vs.SelfDeaf,
			SelfMute: vs.SelfMute,
		}
	}

	return &services.Server{
		ServerID: serverID,
		Name: guild.Name,
		Region: guild.Region,
		OwnerID: guild.OwnerID,
		DefaultChannel: guild.SystemChannelID,
		VoiceStates: voiceStates,
	}, nil
}

func (discord *ClientDiscord) VoiceJoin(serverID, channelID string, muted, deafened bool) (err error) {
	for _, vc := range discord.VCs {
		if vc.GuildID == serverID {
			return vc.ChangeChannel(channelID, muted, deafened)
		}
	}

	vc, err := discord.ChannelVoiceJoin(serverID, channelID, muted, deafened)
	if err != nil {
		return err
	}

	discord.VCs = append(discord.VCs, vc)
	return nil
}
func (discord *ClientDiscord) VoiceLeave(serverID string) (err error) {
	for i := 0; i < len(discord.VCs); i++ {
		if discord.VCs[i].GuildID == serverID {
			if err := discord.VCs[i].Disconnect(); err != nil {
				return err
			}
			discord.VCs = append(discord.VCs[:i], discord.VCs[i+1:]...)
			return nil
		}
	}

	return services.Error("discord: VoiceLeave: unknown VC for server %s", serverID)
}