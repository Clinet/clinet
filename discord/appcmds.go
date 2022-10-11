package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
)

func CmdToAppCommand(cmd *cmds.Cmd) *discordgo.ApplicationCommand {
	appCmd := &discordgo.ApplicationCommand{
		Name: cmd.Aliases[0],
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
			}

			appCmd.Options = append(appCmd.Options, appCmdOpt)
		}
	}

	//Convert subcommands
	if len(cmd.Subcommands) > 0 {
		for _, subCmd := range cmd.Subcommands {
			appCmd.Options = append(appCmd.Options, CmdToAppSubCommand(subCmd))
		}
	}

	return appCmd
}

func CmdToAppSubCommand(cmd *cmds.Cmd) *discordgo.ApplicationCommandOption {
	appSubCmd := &discordgo.ApplicationCommandOption{
		Type: discordgo.ApplicationCommandOptionSubCommand,
		Name: cmd.Aliases[0],
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
			}

			appSubCmd.Options = append(appSubCmd.Options, appCmdOpt)
		}
	}

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