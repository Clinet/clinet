package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func commandSettingsBot(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "prefix":
		if len(args) > 1 {
			if args[1] == botData.CommandPrefix {
				guildSettings[env.Guild.ID].BotPrefix = ""
			} else {
				guildSettings[env.Guild.ID].BotPrefix = args[1]
			}
			return NewGenericEmbed("Bot Settings - Command Prefix", "Successfully set the command prefix to ``"+strings.Replace(args[1], "`", "\\`", -1)+"``.")
		}
		if guildSettings[env.Guild.ID].BotPrefix != "" {
			return NewGenericEmbed("Bot Settings - Command Prefix", "Current command prefix:\n\n"+guildSettings[env.Guild.ID].BotPrefix)
		}
		return NewGenericEmbed("Bot Settings - Command Prefix", "Current command prefix:\n\n"+botData.CommandPrefix)
	}
	return NewErrorEmbed("Bot Settings Error", "Error finding the setting ``"+args[0]+"``.")
}
