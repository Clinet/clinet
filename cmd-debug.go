package main

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

//Debug commands
func commandDebug(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	botData.DebugMode = !botData.DebugMode

	return NewGenericEmbed("Debug Mode", "Debug mode has been set to "+strconv.FormatBool(botData.DebugMode)+".")
}

func commandGCPAuth(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(args) > 0 {
		switch args[0] {
		case "code":
			if len(args) > 1 {
				if err := botData.BotClients.GoogleAssistant.GCPAuth.SetTokenSource(args[1]); err != nil {
					return NewErrorEmbed("GCPAuth Error", "Unable to set the permission code: %v", err)
				}
				return NewGenericEmbed("GCPAuth", "Set the permission code successfully!")
			}
		}
	}
	return nil
}
