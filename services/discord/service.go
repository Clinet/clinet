package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/services"
	"github.com/Clinet/discordgo-embed"
)

var Discord *ClientDiscord

//ClientDiscord implements services.Service and holds a Discord session
type ClientDiscord struct {
	*discordgo.Session

	User *discordgo.User
}

func (discord *ClientDiscord) MsgEdit(msg *services.Message) (ret *services.Message, err error) {
	return nil, nil
}
func (discord *ClientDiscord) MsgRemove(msg *services.Message) (err error) {
	return nil
}
func (discord *ClientDiscord) MsgSend(msg *services.Message) (ret *services.Message, err error) {
	if msg.Context == nil {
		return nil, services.Error("discord: MsgSend(msg: %v): missing context", msg)
	}

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
		return nil, services.Error("discord: MsgSend(msg: %v): unknown MsgContext: %d", msg, msgContext)
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
		case *discordgo.Message:
			discordMsg, err = discord.ChannelMessageSendComplex(msg.ChannelID, &discordgo.MessageSend{Embed: retEmbed.MessageEmbed})
		case *discordgo.Interaction:
			interactionResp := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{retEmbed.MessageEmbed},
				},
			}
			interaction := msg.Context.(*discordgo.Interaction)
			err = discord.InteractionRespond(interaction, interactionResp)
		}
	} else {
		if msg.Content == "" {
			return nil, services.Error("discord: MsgSend(msg: %v): missing content", msg)
		}

		switch msgContext.(type) {
		case *discordgo.Message:
			discordMsg, err = discord.ChannelMessageSend(msg.ChannelID, msg.Content)
		case *discordgo.Interaction:
			interactionResp := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg.Content,
				},
			}
			interaction := msg.Context.(*discordgo.Interaction)
			err = discord.InteractionRespond(interaction, interactionResp)
		}
	}
	if err != nil {
		return nil, err
	}

	ret = msg
	if discordMsg != nil {
		ret.AuthorID = discordMsg.Author.ID
		ret.ServerID = discordMsg.GuildID
	}
	return ret, err
}

func (discord *ClientDiscord) UserBan(user *services.User, reason string, rule int) (msg *services.Message, err error) {
	Log.Trace("Ban(", user.ServerID, ", ", user.UserID, ", ", reason, ", ", rule, ")")
	err = discord.GuildBanCreateWithReason(user.ServerID, user.UserID, reason, 0)
	if err != nil {
		return services.NewMessage().SetContent("Something went wrong while trying to ban them..."), err
	}
	return services.NewMessage().SetContent("And they're gone!"), nil
}
func (discord *ClientDiscord) UserKick(user *services.User, reason string, rule int) (msg *services.Message, err error) {
	Log.Trace("Kick(", user.ServerID, ", ", user.UserID, ", ", reason, ", ", rule, ")")
	err = discord.GuildMemberDeleteWithReason(user.ServerID, user.UserID, reason)
	if err != nil {
		return services.NewMessage().SetContent("Something went wrong while trying to kick them..."), err
	}
	return services.NewMessage().SetContent("And they're gone!"), nil
}