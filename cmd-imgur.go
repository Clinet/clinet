package main

import (
	"github.com/bwmarrin/discordgo"
)

func commandImgur(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	responseEmbed, err := queryImgur(args[0])
	if err != nil {
		return NewErrorEmbed("Imgur Error", "There was an error fetching information about the specified URL.")
	}
	return responseEmbed
}