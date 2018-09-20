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
