package moderation

import (
	"fmt"

	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/logger"
)

//Needed for the cmds framework
var Log *logger.Logger
var Cmds []*cmds.Cmd
var Storage *services.Storage

func Init(log *logger.Logger) error {
	Log = log
	Storage = &services.Storage{}
	if err := Storage.LoadFrom("moderation"); err != nil {
		return err
	}

	Cmds = []*cmds.Cmd{
		cmds.NewCmd("mod", "Provides various moderation utilities", nil).AddSubCmds(
			cmds.NewCmd("ban", "Bans a given user", handleBan).AddArgs(
				cmds.NewCmdArg("user", "Who to actually ban", cmds.ArgTypeUser).SetRequired(),
				cmds.NewCmdArg("reason", "Reason for the ban", "No reason provided."),
				cmds.NewCmdArg("rule", "Rule broken that led to ban", -1),
			),
			cmds.NewCmd("hackban", "Hackbans a given user ID", handleBan).AddArgs(
				cmds.NewCmdArg("user", "ID of the user to ban", "").SetRequired(),
				cmds.NewCmdArg("reason", "Reason for the ban", "No reason provided."),
				cmds.NewCmdArg("rule", "Rule broken that led to ban", -1),
			),
			cmds.NewCmd("kick", "Kicks a given user", handleKick).AddArgs(
				cmds.NewCmdArg("user", "Who to actually kick", cmds.ArgTypeUser).SetRequired(),
				cmds.NewCmdArg("reason", "Reason for the kick", "No reason provided."),
				cmds.NewCmdArg("rule", "Rule broken that led to kick", -1),
			),
			cmds.NewCmd("warn", "Warns a given user", handleWarn).AddArgs(
				cmds.NewCmdArg("user", "Who to actually warn", cmds.ArgTypeUser).SetRequired(),
				cmds.NewCmdArg("reason", "Reason for the warning", "No reason provided."),
				cmds.NewCmdArg("rule", "Rule broken that led to warning", -1),
			),
		),
	}
	return nil
}

func handleBan(ctx *cmds.CmdCtx) *cmds.CmdResp {
	ctxPerms, err := ctx.Service.GetUserPerms(ctx.Server.ServerID, ctx.Channel.ChannelID, ctx.User.UserID)
	if err != nil || !ctxPerms.CanBan() {
		msgErr := services.NewMessage().
			SetContent("You're not allowed to ban anyone!").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	user := ctx.GetArg("user").GetUser()
	user.ServerID = ctx.Server.ServerID //Fill it in for hackbans
	reason := ctx.GetArg("reason").GetString()
	rule := ctx.GetArg("rule").GetInt()

	ban := "You've been banned"
	if rule > 0 {
		ban += " for breaking rule " + fmt.Sprintf("%d", rule)
	}
	ban += "."
	if reason != "" {
		ban += " The following reason was given: " + reason
	}

	msgBan := services.NewMessage().
		SetContent(ban).
		SetColor(0xFF0000)

	server, err := ctx.Service.GetServer(ctx.Server.ServerID)
	if err != nil {
		Log.Error(err)
	} else {
		msgBan.SetTitle(server.Name)
	}

	//Ship off the DM
	msgBan.ChannelID = user.UserID
	if _, err = ctx.Service.MsgSend(msgBan); err != nil {
		Log.Error(err)
	}

	if err := ctx.Service.UserBan(user, reason, rule); err != nil {
		Log.Error(err)
		msgErr := services.NewMessage().
			SetContent("Something went wrong while trying to ban that user...").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	msg := services.NewMessage().
		SetContent("You banished them to the shadow realm!").
		SetColor(0x1C1C1C)
	return cmds.CmdRespFromMsg(msg).SetReady(true)
}
func handleKick(ctx *cmds.CmdCtx) *cmds.CmdResp {
	ctxPerms, err := ctx.Service.GetUserPerms(ctx.Server.ServerID, ctx.Channel.ChannelID, ctx.User.UserID)
	if err != nil || !ctxPerms.CanKick() {
		msgErr := services.NewMessage().
			SetContent("You're not allowed to kick anyone!").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	user := ctx.GetArg("user").GetUser()
	reason := ctx.GetArg("reason").GetString()
	rule := ctx.GetArg("rule").GetInt()

	kick := "You've been kicked"
	if rule > 0 {
		kick += " for breaking rule " + fmt.Sprintf("%d", rule)
	}
	kick += "."
	if reason != "" {
		kick += " The following reason was given: " + reason
	}

	msgKick := services.NewMessage().
		SetContent(kick).
		SetColor(0xCCCC09) //Dirty yellow?

	server, err := ctx.Service.GetServer(ctx.Server.ServerID)
	if err != nil {
		Log.Error(err)
	} else {
		msgKick.SetTitle(server.Name)
	}

	//Ship off the DM
	msgKick.ChannelID = user.UserID
	if _, err = ctx.Service.MsgSend(msgKick); err != nil {
		Log.Error(err)
	}

	if err := ctx.Service.UserKick(user, reason, rule); err != nil {
		Log.Error(err)
		msgErr := services.NewMessage().
			SetContent("Something went wrong while trying to kick that user...").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	msg := services.NewMessage().
		SetContent("You kicked them over the gates!").
		SetColor(0x1C1C1C)
	return cmds.CmdRespFromMsg(msg).SetReady(true)
}

type Warning struct {
	Reason string
	Rule   int
}
func handleWarn(ctx *cmds.CmdCtx) *cmds.CmdResp {
	ctxPerms, err := ctx.Service.GetUserPerms(ctx.Server.ServerID, ctx.Channel.ChannelID, ctx.User.UserID)
	//Only users with kick or ban perms can warn someone, with the side effect of automatic ban/kick failing if the warner can't do so
	if err != nil || !ctxPerms.CanKick() || !ctxPerms.CanBan() {
		msgErr := services.NewMessage().
			SetContent("You're not allowed to warn anyone!").
			SetColor(0xFF0000)
		return cmds.CmdRespFromMsg(msgErr).SetReady(true)
	}

	user := ctx.GetArg("user").GetUser()
	reason := ctx.GetArg("reason").GetString()
	rule := ctx.GetArg("rule").GetInt()

	warnings := make([]*Warning, 0)
	rawWarnings, err := Storage.UserGet(user.UserID, "warnings")
	if err == nil {
		warnings = rawWarnings.([]*Warning)
	}
	warnings = append(warnings, &Warning{Reason: reason, Rule: rule})
	Storage.UserSet(user.UserID, "warnings", warnings)

	warnLimit := 3
	rawWarnLimit, err := Storage.ServerGet(ctx.Server.ServerID, "warnLimit")
	if err != nil {
		Storage.ServerSet(ctx.Server.ServerID, "warnLimit", warnLimit)
	} else {
		warnLimit = rawWarnLimit.(int)
	}

	if len(warnings) >= warnLimit {
		shouldBan := false
		rawShouldBan, err := Storage.ServerGet(ctx.Server.ServerID, "warnLimitShouldBan")
		if err != nil {
			Storage.ServerSet(ctx.Server.ServerID, "warnLimitShouldBan", shouldBan)
		} else {
			shouldBan = rawShouldBan.(bool)
		}

		//Conveniently, commands may simply call each other now
		if shouldBan {
			return handleBan(ctx)
		}
		return handleKick(ctx)
	} else {
		warning := "You've been warned"
		if rule > 0 {
			warning += " for breaking rule " + fmt.Sprintf("%d", rule)
		}
		warning += "."
		if reason != "" {
			warning += " The following reason was given: " + reason
		}

		msgWarning := services.NewMessage().
			SetContent(warning).
			SetColor(0xCCCC09) //Dirty yellow?

		server, err := ctx.Service.GetServer(ctx.Server.ServerID)
		if err != nil {
			Log.Error(err)
		} else {
			msgWarning.SetTitle(server.Name)
		}

		//Ship off the DM
		msgWarning.ChannelID = user.UserID
		if _, err = ctx.Service.MsgSend(msgWarning); err != nil {
			Log.Error(err)
			msgErr := services.NewMessage().
				SetContent("Something went wrong while trying to DM the warning, but it applied anyway.").
				SetColor(0xCCCC09)
			return cmds.CmdRespFromMsg(msgErr).SetReady(true)
		}
	}

	return cmds.NewCmdRespMsg("The user has been warned!")
}