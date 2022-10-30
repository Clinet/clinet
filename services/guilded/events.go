package guilded

import (
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/guildrone"
)

func guildedReady(session *guildrone.Session, event *guildrone.Ready) {
	Log.Trace("--- guildedReady(", event, ") ---")
	for Guilded == nil {
		//Wait for Discord to finish connecting on our end
		if Guilded != nil {
			break
		}
	}
	Guilded.User = event.User
	Log.Info("Logged into Guilded as ", Guilded.User, "!")
}

func guildedChatMessageCreated(session *guildrone.Session, event *guildrone.ChatMessageCreated) {
	//Log.Trace("--- guildedChatMessageCreated(", event, ") ---", event.Message.ID, event.ServerID, event.Message.ChannelID, event.Message.CreatedBy)
	if event == nil || event.Message.Content == "" {
		return
	}

	msg := &services.Message{
		UserID: event.Message.CreatedBy,
		MessageID: event.Message.ID,
		ChannelID: event.Message.ChannelID,
		ServerID: event.Message.ServerID,
		Content: event.Message.Content,
		Context: event.Message,
	}

	cmdName, cmdResps, err := cmds.CmdHandler(msg, Guilded)
	if err != nil {
		Log.Error(err)
		msgErr := msg
		msgErr.Content = err.Error()
		Guilded.MsgSend(msgErr)
		return
	}

	if len(cmdResps) == 0 {
		Log.Warn("no cmdresp")
		return
	}
	for i := 0; i < len(cmdResps); i++ {
		if cmdResps[i] == nil {
			Log.Warn("ignoring empty cmdresp")
			continue
		}

		cmdResps[i].OnReady(func(r *cmds.CmdResp) {
			Log.Trace("Response to message for command " + cmdName + ": " + r.String())
			r.Context = event.Message
			r.ChannelID = event.Message.ChannelID

			msg, err := Guilded.MsgSend(r.Message)
			if err != nil {
				Log.Error(err)
				return
			}
			Log.Trace("Sent message: ", msg)
		})
	}
}