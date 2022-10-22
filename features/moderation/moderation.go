package moderation

import (
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/logger"
)

//Needed for the cmds framework
var Log *logger.Logger
var CmdRoot *cmds.Cmd
var Storage *services.Storage

func Init(log *logger.Logger) error {
	Log = log
	Storage = &services.Storage{}
	if err := Storage.LoadFrom("moderation"); err != nil {
		return err
	}

	CmdRoot = cmds.NewCmd("mod", "Provides various moderation utilities", nil).AddSubCmds(
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
	)
	return nil
}

func handleBan(ctx *cmds.CmdCtx) *cmds.CmdResp {
	user := ctx.GetArg("user").GetUser()
	user.ServerID = ctx.Server.ServerID //Fill it in for hackban
	reason := ctx.GetArg("reason").GetString()
	rule := ctx.GetArg("rule").GetInt()

	msg, err := ctx.Service.UserBan(user, reason, rule)
	if err != nil {
		Log.Error(err)
	}

	return cmds.CmdRespFromMsg(msg).SetColor(0x1C1C1C).SetReady(true)
}
func handleKick(ctx *cmds.CmdCtx) *cmds.CmdResp {
	user := ctx.GetArg("user").GetUser()
	reason := ctx.GetArg("reason").GetString()
	rule := ctx.GetArg("rule").GetInt()
	
	msg, err := ctx.Service.UserKick(user, reason, rule)
	if err != nil {
		Log.Error(err)
	}

	return cmds.CmdRespFromMsg(msg).SetColor(0x1C1C1C).SetReady(true)
}

type Warning struct {
	Reason string
	Rule   int
}
func handleWarn(ctx *cmds.CmdCtx) *cmds.CmdResp {
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
		msg, err := ctx.Service.UserKick(user, reason, rule)
		if err != nil {
			Log.Error(err)
		}
		return cmds.CmdRespFromMsg(msg).SetColor(0x1C1C1C).SetReady(true)
	}

	//TODO: DM user

	return cmds.NewCmdRespMsg("The user has been warned!")
}