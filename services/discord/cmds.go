package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/services"
)

func cmdHandler(cmd *cmds.Cmd, interaction *discordgo.Interaction, eventOpts []*discordgo.ApplicationCommandInteractionDataOption, subCmd bool) (string, []*cmds.CmdResp) {
	cmdAlias := interaction.ApplicationCommandData().Name

	if !subCmd && len(eventOpts) == 1 && eventOpts[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		Log.Trace("Checking subcommands for " + cmd.Name + " using subcommand " + eventOpts[0].Name)
		for i := 0; i < len(cmd.Subcommands); i++ {
			Log.Trace("- " + cmd.Subcommands[i].Name + " == " + eventOpts[0].Name)
			if cmd.Subcommands[i].Name == eventOpts[0].Name {
				Log.Trace("> Testing subcommand " + cmd.Subcommands[i].Name)
				return cmdHandler(cmd.Subcommands[i], interaction, eventOpts[0].Options, true)
			}
		}
		Log.Error("Command " + cmd.Name + " has no subcommand " + eventOpts[0].Name)
		return cmdAlias, nil
	}

	user := &services.User{
		ServerID: interaction.GuildID,
		UserID: interaction.Member.User.ID,
	}
	channel := &services.Channel{
		ServerID: interaction.GuildID,
		ChannelID: interaction.ChannelID,
	}
	server := &services.Server{
		ServerID: interaction.GuildID,
	}
	message := &services.Message{
		MessageID: interaction.ID,
	}

	cmdCtx := cmds.NewCmdCtx().
		SetAlias(cmdAlias).
		SetUser(user).
		SetChannel(channel).
		SetServer(server).
		SetMessage(message).
		SetService(Discord)
	cmdArgs := discordCmdArgs(cmd, cmdCtx, eventOpts)
	cmdCtx.AddArgs(cmdArgs...)
	cmdBuilder := &cmds.CmdBuilderCommand{Command: cmd, Context: cmdCtx} //Build up a command builder
	cmdRuntime := cmds.CmdBatch(cmdBuilder) //Prepare a command runtime with just this command (but can be supplied additional cmdBuilders)

	cmdResps := cmdRuntime.Run() //Run the commands and return their responses
	if len(cmdResps) == 0 {
		Log.Warn("No responses for cmd " + cmdAlias)
	}
	return cmdAlias, cmdResps
}

func discordCmdArgs(cmd *cmds.Cmd, cmdCtx *cmds.CmdCtx, eventArgs []*discordgo.ApplicationCommandInteractionDataOption) []*cmds.CmdArg {
	cmdArgs := cmd.Args
	finalArgs := make([]*cmds.CmdArg, 0)

	for i := 0; i < len(cmdArgs); i++ {
		foundArg := false
		for j := 0; j < len(eventArgs); j++ {
			if cmdArgs[i].Name == eventArgs[j].Name {
				cmdArg := cmdArgs[i]
				switch eventArgs[j].Type {
				case discordgo.ApplicationCommandOptionString:
					cmdArg.Value = eventArgs[j].StringValue()
				case discordgo.ApplicationCommandOptionInteger:
					cmdArg.Value = eventArgs[j].IntValue()
				case discordgo.ApplicationCommandOptionUser:
					user := eventArgs[j].UserValue(Discord.Session)
					cmdArg.Value = &services.User{
						ServerID: cmdCtx.Server.ServerID,
						UserID: user.ID,
					}
				default:
					cmdArg.Value = eventArgs[j].Value
				}
				Log.Trace("Filled in cmdArg: ", cmdArg)
				finalArgs = append(finalArgs, cmdArg)
				foundArg = true
				break
			}
		}
		if !foundArg {
			Log.Trace("Failed to find cmdArg: ", cmdArgs[i])
			finalArgs = append(finalArgs, cmdArgs[i])
		}
	}

	return finalArgs
}