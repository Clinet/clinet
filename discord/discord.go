package discord

import (
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
	"github.com/JoshuaDoes/logger"
)

var Log *logger.Logger

//DiscordSession holds a Discord session
type DiscordSession struct {
	*discordgo.Session
}

var DiscordCfg *CfgDiscord

//Configuration for Discord sessions
type CfgDiscord struct {
	//Stuff for communication with Discord
	Token string `json:"token"`

	//Trust for Discord communication
	DisplayName   string `json:"displayName"`   //The display name for communicating on Discord
	OwnerID       string `json:"ownerID"`       //The user ID of the bot owner on Discord
	CommandPrefix string `json:"commandPrefix"` //The command prefix to use when invoking the bot on Discord
}

//StartDiscord returns a connected DiscordSession, up to caller to close it
func StartDiscord(cfg *CfgDiscord) *DiscordSession {
	DiscordCfg = cfg
	Log.Trace("--- StartDiscord(" + DiscordCfg.Token + ") ---") //Maybe I should sensor the bot token? Protect your logs and your config

	Log.Debug("Creating Discord struct")
	Discord, err := discordgo.New("Bot " + DiscordCfg.Token)
	if err != nil {
		Log.Fatal("Unable to connect to Discord!")
	}

	//Only enable informational Discord logging if we're tracing
	if Log.Verbosity == 2 {
		Log.Debug("Setting Discord log level to informational")
		Discord.LogLevel = discordgo.LogInformational
	}

	Log.Info("Registering Discord event handlers")
	Discord.AddHandler(discordReady)
	Discord.AddHandler(discordChannelCreate)
	Discord.AddHandler(discordChannelUpdate)
	Discord.AddHandler(discordChannelDelete)
	Discord.AddHandler(discordGuildUpdate)
	Discord.AddHandler(discordGuildBanAdd)
	Discord.AddHandler(discordGuildBanRemove)
	Discord.AddHandler(discordGuildMemberAdd)
	Discord.AddHandler(discordGuildMemberRemove)
	Discord.AddHandler(discordGuildRoleCreate)
	Discord.AddHandler(discordGuildRoleUpdate)
	Discord.AddHandler(discordGuildRoleDelete)
	Discord.AddHandler(discordGuildEmojisUpdate)
	Discord.AddHandler(discordUserUpdate)
	Discord.AddHandler(discordVoiceStateUpdate)
	Discord.AddHandler(discordMessageCreate)
	Discord.AddHandler(discordMessageDelete)
	Discord.AddHandler(discordMessageDeleteBulk)
	Discord.AddHandler(discordMessageUpdate)
	Discord.AddHandler(discordMessageReactionAdd)
	Discord.AddHandler(discordMessageReactionRemove)
	Discord.AddHandler(discordMessageReactionRemoveAll)

	Log.Info("Connecting to Discord")
	err = Discord.Open()
	if err != nil {
		Log.Fatal("Unable to connect to Discord!", err)
		return nil
	}

	Log.Info("Connected to Discord!")
	return &DiscordSession{Discord}
}

func cmdDiscord(session *discordgo.Session, message *discordgo.Message) ([]*cmds.CmdResp, error) {
	if message == nil {
		return nil, cmds.ErrCmdEmptyMsg
	}
	if message.Author.Bot {
		return nil, nil
	}
	content := message.Content
	if content == "" {
		return nil, nil
	}
	if content[:len(DiscordCfg.CommandPrefix)] != DiscordCfg.CommandPrefix {
		return nil, nil
	}
	content = content[len(DiscordCfg.CommandPrefix):]

	//Prepare cmd context
	contentDisplay, err := message.ContentWithMoreMentionsReplaced(session)
	if err != nil {
		return nil, err
	}
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		return nil, err
	}
	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		return nil, err
	}
	author := message.Author
	member, err := session.GuildMember(guild.ID, author.ID)
	if err != nil {
		return nil, err
	}

	//Build cmd context
	ctx := &cmds.CmdCtx{
		Content: content,
		ContentDisplay: contentDisplay,
		CmdPrefix: DiscordCfg.CommandPrefix,
		CmdEdited: false,
		Server: guild,
		Channel: channel,
		Message: message,
		User: author,
		Member: member,
	}

	//Build cmd list
	cmdList := make([]*cmds.CmdBuilderCommand, 0)
	cmd, err := cmds.CmdMessage(ctx, content)
	if err != nil {
		return nil, err
	}
	cmdList = append(cmdList, cmd)

	//Build cmd builder
	cmdRuntime := cmds.CmdBatch(cmdList...)

	//Run cmds and get their responses
	cmdResps := cmdRuntime.Run()
	if len(cmdResps) == 0 {
		Log.Error("no responses")
		return nil, cmds.ErrCmdNoResp
	}

	//Return their responses
	return cmdResps, nil
}

func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	Log.Trace("--- discordReady(", event, ") ---")
}

func discordChannelCreate(session *discordgo.Session, event *discordgo.ChannelCreate) {
	Log.Trace("--- discordChannelCreate(", event, ") ---")
}

func discordChannelUpdate(session *discordgo.Session, event *discordgo.ChannelUpdate) {
	Log.Trace("--- discordChannelUpdate(", event, ") ---")
}

func discordChannelDelete(session *discordgo.Session, event *discordgo.ChannelDelete) {
	Log.Trace("--- discordChannelDelete(", event, ") ---")
}

func discordGuildUpdate(session *discordgo.Session, event *discordgo.GuildUpdate) {
	Log.Trace("--- discordGuildUpdate(", event, ") ---")
}

func discordGuildBanAdd(session *discordgo.Session, event *discordgo.GuildBanAdd) {
	Log.Trace("--- discordGuildBanAdd(", event, ") ---")
}

func discordGuildBanRemove(session *discordgo.Session, event *discordgo.GuildBanRemove) {
	Log.Trace("--- discordGuildBanRemove(", event, ") ---")
}

func discordGuildMemberAdd(session *discordgo.Session, event *discordgo.GuildMemberAdd) {
	Log.Trace("--- discordGuildMemberAdd(", event, ") ---")
}

func discordGuildMemberRemove(session *discordgo.Session, event *discordgo.GuildMemberRemove) {
	Log.Trace("--- discordGuildMemberRemove(", event, ") ---")
}

func discordGuildRoleCreate(session *discordgo.Session, event *discordgo.GuildRoleCreate) {
	Log.Trace("--- discordGuildRoleCreate(", event, ") ---")
}

func discordGuildRoleUpdate(session *discordgo.Session, event *discordgo.GuildRoleUpdate) {
	Log.Trace("--- discordGuildRoleUpdate(", event, ") ---")
}

func discordGuildRoleDelete(session *discordgo.Session, event *discordgo.GuildRoleDelete) {
	Log.Trace("--- discordGuildRoleDelete(", event, ") ---")
}

func discordGuildEmojisUpdate(session *discordgo.Session, event *discordgo.GuildEmojisUpdate) {
	Log.Trace("--- discordGuildEmojisUpdate(", event, ") ---")
}

func discordUserUpdate(session *discordgo.Session, event *discordgo.UserUpdate) {
	Log.Trace("--- discordUserUpdate(", event, ") ---")
}

func discordVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	Log.Trace("--- discordVoiceStateUpdate(", event, ") ---")
}

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	Log.Trace("--- discordMessageCreate(", event, ") ---")
	message, err := session.ChannelMessage(event.ChannelID, event.ID)
	if err != nil {
		Log.Error(message, err)
	}

	resps, err := cmdDiscord(session, message)
	if err != nil {
		Log.Error(resps, err)
		return
	}

	for _, resp := range resps {
		random := rand.Intn(len(resp.Messages))
		message := resp.Messages[random]

		respMessage, err := session.ChannelMessageSend(event.ChannelID, message)
		if err != nil {
			Log.Error(respMessage, err)
		}
	}
}

func discordMessageDelete(session *discordgo.Session, event *discordgo.MessageDelete) {
	Log.Trace("--- discordMessageDelete(", event, ") ---")
}

func discordMessageDeleteBulk(session *discordgo.Session, event *discordgo.MessageDeleteBulk) {
	Log.Trace("--- discordMessageDeleteBulkage(", event, ") ---")
}

func discordMessageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	Log.Trace("--- discordMessageUpdate(", event, ") ---")
	/*message, err := session.ChannelMessage(event.ChannelID, event.ID)
	if err != nil {
		Log.Error(message, err)
	}

	resps, err := cmdDiscord(session, message)
	if err != nil {
		Log.Error(resps, err)
		return
	}*/

	//update messages
}

func discordMessageReactionAdd(session *discordgo.Session, event *discordgo.MessageReactionAdd) {
	Log.Trace("--- discordMessageReactionAdd(", event, ") ---")
}

func discordMessageReactionRemove(session *discordgo.Session, event *discordgo.MessageReactionRemove) {
	Log.Trace("--- discordMessageReactionRemove(", event, ") ---")
}

func discordMessageReactionRemoveAll(session *discordgo.Session, event *discordgo.MessageReactionRemoveAll) {
	Log.Trace("--- discordMessageReactionRemoveAll(", event, ") ---")
}
