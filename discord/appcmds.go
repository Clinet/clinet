package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/services"
)

func CmdToAppCommand(cmd *cmds.Cmd) *discordgo.ApplicationCommand {
	appCmd := &discordgo.ApplicationCommand{
		Name: cmd.Name,
		Description: cmd.Description,
	}

	//Convert command arguments
	if len(cmd.Args) > 0 {
		for _, arg := range cmd.Args {
			appCmdOpt := &discordgo.ApplicationCommandOption{
				Name: arg.Name,
				Description: arg.Description,
				Required: arg.Required,
			}

			switch arg.Value.(type) {
			case string:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionString
			case int:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionInteger
			case bool:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionBoolean
			case *services.User:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionUser
			}

			Log.Trace("- Built appCmdOpt: ", appCmdOpt)
			appCmd.Options = append(appCmd.Options, appCmdOpt)
		}
	}

	//Convert subcommands
	if len(cmd.Subcommands) > 0 {
		for _, subCmd := range cmd.Subcommands {
			appCmd.Options = append(appCmd.Options, CmdToAppSubCommand(subCmd))
		}
	}

	Log.Trace("- Built cmd: ", appCmd)
	return appCmd
}

func CmdToAppSubCommand(cmd *cmds.Cmd) *discordgo.ApplicationCommandOption {
	appSubCmd := &discordgo.ApplicationCommandOption{
		Type: discordgo.ApplicationCommandOptionSubCommand,
		Name: cmd.Name,
		Description: cmd.Description,
	}

	//Convert command arguments
	if len(cmd.Args) > 0 {
		for _, arg := range cmd.Args {
			appCmdOpt := &discordgo.ApplicationCommandOption{
				Name: arg.Name,
				Description: arg.Description,
				Required: arg.Required,
			}

			switch arg.Value.(type) {
			case string:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionString
			case int:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionInteger
			case bool:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionBoolean
			case *services.User:
				appCmdOpt.Type = discordgo.ApplicationCommandOptionUser
			}

			Log.Trace("- Built appSubCmdOpt: ", appCmdOpt)
			appSubCmd.Options = append(appSubCmd.Options, appCmdOpt)
		}
	}

	Log.Trace("- Built subcmd: ", appSubCmd)
	return appSubCmd
}

func CmdsToAppCommands() []*discordgo.ApplicationCommand {
	appCmds := make([]*discordgo.ApplicationCommand, 0)
	
	for _, cmd := range cmds.Commands {
		//Add finalized cmd to list
		appCmds = append(appCmds, CmdToAppCommand(cmd))
	}
	
	return appCmds
}