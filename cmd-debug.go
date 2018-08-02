package main

import (
	"strconv"
	"strings"
	
	"github.com/bwmarrin/discordgo"
)

//Debug commands
func commandBotInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	botGuilds, err := botData.DiscordSession.UserGuilds(100, "", "")
	if err != nil {
		return NewErrorEmbed("Bot Info Error", "An error occurred retrieving a list of guilds.")
	}
	botGuildNames := make([]string, 0)
	for i := 0; i < len(botGuilds); i++ {
		botGuildNames = append(botGuildNames, botGuilds[i].Name)
	}

	return NewEmbed().
		SetTitle("Bot Info").
		SetDescription("Info regarding this bot.").
		AddField("Guild List ("+strconv.Itoa(len(botGuildNames))+")", strings.Join(botGuildNames, "\n")).
		SetColor(0x1C1C1C).MessageEmbed
}
