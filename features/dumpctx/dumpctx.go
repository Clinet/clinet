package dumpctx

import (
	"github.com/Clinet/clinet/cmds"
)

var CmdRoot *cmds.Cmd

func init() {
	CmdRoot = cmds.NewCmd("dumpctx", "DEBUG: Dumps command context as built by Clinet", nil)
	CmdRoot.AddSubCmds(
		cmds.NewCmd("sub1", "Test subcommand 1", handlerDumpCtx),
		cmds.NewCmd("sub2", "Test subcommand 2", handlerDumpCtx),
	)
}

func handlerDumpCtx(ctx *cmds.CmdCtx) *cmds.CmdResp {
	return cmds.NewCmdRespEmbed("Dump of ctx (*cmds.CmdCtx)", "```JSON\n" + ctx.String() + "```").SetColor(0x1C1C1C)
}