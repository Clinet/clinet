package guilded

import (
	"fmt"

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

func (guilded *ClientGuilded) Shutdown() {
	_ = guilded.Close()
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

	isPrivate := false
	if msg.ServerID == "" {
		isPrivate = true
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

		guildedMsg, err = guilded.ChannelMessageCreateComplex(msg.ChannelID, &guildrone.MessageCreate{IsPrivate: isPrivate, Embeds: []guildrone.ChatEmbed{retEmbed}})
	} else {
		if msg.Content == "" {
			return nil, services.Error("guilded: MsgSend(msg: %v): missing content", msg)
		}

		guildedMsg, err = guilded.ChannelMessageCreateComplex(msg.ChannelID, &guildrone.MessageCreate{IsPrivate: isPrivate, Content: msg.Content})
	}
	if err != nil {
		return nil, err
	}

	if guildedMsg != nil {
		ret = &services.Message{
			UserID: guildedMsg.CreatedBy,
			MessageID: guildedMsg.ID,
			ChannelID: guildedMsg.ChannelID,
			ServerID: guildedMsg.ServerID,
			Content: guildedMsg.Content,
			Context: guildedMsg,
		}
	}
	return ret, err
}

func (guilded *ClientGuilded) GetUser(serverID, userID string) (ret *services.User, err error) {
	user, err := guilded.ServerMemberGet(serverID, userID)
	if err != nil {
		return nil, err
	}
	userRoles := make([]*services.Role, len(user.RoleIds))
	for i := 0; i < len(userRoles); i++ {
		userRoles[i] = &services.Role{
			RoleID: fmt.Sprintf("%d", user.RoleIds[i]),
		}
	}
	return &services.User{
		ServerID: serverID,
		UserID: userID,
		Username: user.User.Name,
		Nickname: user.Nickname,
		Roles: userRoles,
	}, nil
}
func (guilded *ClientGuilded) GetUserPerms(serverID, channelID, userID string) (perms *services.Perms, err error) {
	server, err := guilded.GetServer(serverID)
	if err != nil {
		return nil, err
	}

	perms = &services.Perms{}
	//TODO: Permission mapping from Guilded

	if server.OwnerID == userID {
		perms.Administrator = true
	}

	return perms, nil
}
func (guilded *ClientGuilded) UserBan(user *services.User, reason string, rule int) (err error) {
	Log.Trace("Ban(", user.ServerID, ", ", user.UserID, ", ", reason, ", ", rule, ")")
	_, err = guilded.ServerMemberBanCreate(user.ServerID, user.UserID, reason)
	return err
}
func (guilded *ClientGuilded) UserKick(user *services.User, reason string, rule int) (err error) {
	Log.Trace("Kick(", user.ServerID, ", ", user.UserID, ", ", reason, ", ", rule, ")")
	return guilded.ServerMemberKick(user.ServerID, user.UserID)
}

func (guilded *ClientGuilded) GetServer(serverID string) (server *services.Server, err error) {
	srv, err := guilded.ServerGet(serverID)
	if err != nil {
		return nil, err
	}
	return &services.Server{
		ServerID: serverID,
		Name: srv.Name,
		Region: srv.Timezone,
		OwnerID: srv.OwnerID,
		DefaultChannel: srv.DefaultChannelID,
		VoiceStates: make([]*services.VoiceState, 0), //TODO: Voice states from Guilded
	}, nil
}

func (guilded *ClientGuilded) VoiceJoin(serverID, channelID string, muted, deafened bool) (err error) {
	return services.Error("guilded: VoiceJoin: stub")
}
func (guilded *ClientGuilded) VoiceLeave(serverID string) (err error) {
	return services.Error("guilded: VoiceLeave: stub")
}