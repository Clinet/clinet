package guilded

import (
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/guildrone"
	"github.com/JoshuaDoes/logger"
)

var Log *logger.Logger
var Guilded *ClientGuilded

//ClientGuilded implements services.Service and holds a Guilded session
type ClientGuilded struct {
	*guildrone.Session

	cmdPrefix string
	User      guildrone.BotUser
}

func (guilded *ClientGuilded) CmdPrefix() string {
	return guilded.cmdPrefix
}

func (guilded *ClientGuilded) Login(cfg interface{}) (err error) {
	GuildedCfg = cfg.(*CfgGuilded)
	Log.Trace("--- StartGuilded() ----")

	Log.Debug("Creating Guilded struct...")
	guildedClient, err := guildrone.New(GuildedCfg.Token)
	if err != nil {
		return err
	}

	Log.Info("Registering Guilded event handlers...")
	guildedClient.AddHandler(guildedReady)
	guildedClient.AddHandler(guildedChatMessageCreated)

	Log.Info("Connecting to Guilded...")
	err = guildedClient.Open()
	if err != nil {
		return err
	}

	Log.Info("Connected to Guilded!")
	Guilded = &ClientGuilded{guildedClient, GuildedCfg.CmdPrefix, guildrone.BotUser{}}
	return nil
}

func (guilded *ClientGuilded) MsgEdit(msg *services.Message) (ret *services.Message, err error) {
	return nil, nil
}
func (guilded *ClientGuilded) MsgRemove(msg *services.Message) (err error) {
	return nil
}
func (guilded *ClientGuilded) MsgSend(msg *services.Message) (ret *services.Message, err error) {
	if msg.ChannelID == "" {
		return nil, services.Error("guilded: MsgSend(msg: %v): missing channel ID", msg)
	}

	var guildedMsg *guildrone.ChatMessage
	if msg.Title != "" || msg.Color != nil || msg.Image != "" {
		retEmbed := guildrone.ChatEmbed{Description: msg.Content}
		if msg.Title != "" {
			retEmbed.Title = msg.Title
		}
		if msg.Color != nil {
			retEmbed.Color = *msg.Color
		}
		if msg.Image != "" {
			retEmbed.Image = &guildrone.ChatEmbedImage{
				URL: msg.Image,
			}
		}

		guildedMsg, err = guilded.ChannelMessageCreateComplex(msg.ChannelID, &guildrone.MessageCreate{Embeds: []guildrone.ChatEmbed{retEmbed}})
	} else {
		if msg.Content == "" {
			return nil, services.Error("guilded: MsgSend(msg: %v): missing content", msg)
		}

		guildedMsg, err = guilded.ChannelMessageCreate(msg.ChannelID, msg.Content)
	}
	if err != nil {
		return nil, err
	}

	if guildedMsg != nil {
		ret = &services.Message{
			AuthorID: guildedMsg.CreatedBy,
			MessageID: guildedMsg.ID,
			ChannelID: guildedMsg.ChannelID,
			ServerID: guildedMsg.ServerID,
			Content: guildedMsg.Content,
			Context: guildedMsg,
		}
	}
	return ret, err
}

func (guilded *ClientGuilded) UserBan(user *services.User, reason string, rule int) (msg *services.Message, err error) {
	return nil, nil
}
func (guilded *ClientGuilded) UserKick(user *services.User, reason string, rule int) (msg *services.Message, err error) {
	return nil, nil
}