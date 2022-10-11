package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
)

func cmdHandler(cmd *cmds.Cmd, cmdAlias string, eventOpts []*discordgo.ApplicationCommandInteractionDataOption) (string, []*cmds.CmdResp) {
	if len(eventOpts) == 1 && eventOpts[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		for i := 0; i < len(cmd.Subcommands); i++ {
			if cmd.Subcommands[i].Name == eventOpts[0].Name {
				return cmdHandler(cmd.Subcommands[i], eventOpts[0].Name, eventOpts[0].Options)
			}
		}
		Log.Error("Command " + cmd.Name + " has no subcommand " + eventOpts[0].Name)
		return cmdAlias, nil
	}

	cmdArgs := discordCmdArgs(cmd, eventOpts)
	cmdCtx := cmds.NewCmdCtx().
		SetAlias(cmdAlias).
		AddArgs(cmdArgs...).
		SetService(Discord)
	cmdBuilder := &cmds.CmdBuilderCommand{Command: cmd, Context: cmdCtx} //Build up a command builder
	cmdRuntime := cmds.CmdBatch(cmdBuilder) //Prepare a command runtime with just this command (but can be supplied additional cmdBuilders)

	cmdResps := cmdRuntime.Run() //Run the commands and return their responses
	if len(cmdResps) == 0 {
		Log.Warn("No responses for cmd " + cmdAlias)
	}
	return cmdAlias, cmdResps
}

func discordCmdArgs(cmd *cmds.Cmd, eventArgs []*discordgo.ApplicationCommandInteractionDataOption) []*cmds.CmdArg {
	cmdArgs := cmd.Args
	finalArgs := make([]*cmds.CmdArg, 0)

	for i := 0; i < len(cmdArgs); i++ {
		foundArg := false
		for j := 0; j < len(eventArgs); j++ {
			if cmdArgs[i].Name == eventArgs[j].Name {
				cmdArg := &cmdArgs[i]
				switch eventArgs[j].Type {
				case discordgo.ApplicationCommandOptionString:
					cmdArg.Value = eventArgs[j].StringValue()
				default:
					cmdArg.Value = eventArgs[j].Value
				}
				finalArgs = append(finalArgs, cmdArg)
				foundArg = true
				break
			}
		}
		if !foundArg {
			finalArgs = append(finalArgs, &cmdArgs[i])
		}
	}

	return finalArgs
}