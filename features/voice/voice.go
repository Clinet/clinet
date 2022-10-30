package voice

import (
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/logger"
)

var Log *logger.Logger
var Cmds []*cmds.Cmd
var Storage *services.Storage

func Init() error {
	Storage = &services.Storage{}
	if err := Storage.LoadFrom("voice"); err != nil {
		Log.Error(err)
		return err
	}

	Cmds = []*cmds.Cmd{
		cmds.NewCmd("voice", "For more direct control of voice channels", nil).AddSubCmds(
			cmds.NewCmd("join", "Joins your active voice channel", handleJoin),
			cmds.NewCmd("leave", "Leaves your active voice channel", handleLeave),
		),
	}
	return nil
}

func handleJoin(ctx *cmds.CmdCtx) *cmds.CmdResp {
	server, err := ctx.Service.GetServer(ctx.Server.ServerID)
	if err != nil {
		Log.Error(err)
		msgErr := services.NewMessage().
			SetContent("We can't find your server!").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	for i := 0; i < len(server.VoiceStates); i++ {
		vs := server.VoiceStates[i]
		if vs.UserID == ctx.User.UserID {
			if err := ctx.Service.VoiceJoin(ctx.Server.ServerID, vs.ChannelID, true, true); err != nil {
				Log.Error(err)
				msgErr := services.NewMessage().
					SetContent("We were unable to join your voice channel!").
					SetColor(0xFF0000)
				return cmds.CmdRespFromMsg(msgErr).SetReady(true)
			}

			msg := services.NewMessage().
				SetContent("I'm in your voice channel now!").
				SetColor(0x1C1C1C)
			return cmds.CmdRespFromMsg(msg).SetReady(true)
		}
	}

	msgErr := services.NewMessage().
		SetContent("We can't find your voice channel!").
		SetColor(0xFF0000)
	return cmds.CmdRespFromMsg(msgErr).SetReady(true)
}
func handleLeave(ctx *cmds.CmdCtx) *cmds.CmdResp {
	if err := ctx.Service.VoiceLeave(ctx.Server.ServerID); err != nil {
		Log.Error(err)
		msgErr := services.NewMessage().
			SetContent("We were unable to leave your voice channel!").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	msg := services.NewMessage().
		SetContent("I left your voice channel!").
		SetColor(0x1C1C1C)
	return cmds.CmdRespFromMsg(msg).SetReady(true)
}