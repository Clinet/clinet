package main

import (
	"github.com/JoshuaDoes/go-cve"
	"github.com/bwmarrin/discordgo"
)

func commandCVE(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	cveData, err := cve.GetCVE(args[0])
	if err != nil {
		return NewErrorEmbed("CVE Error", "There was an error fetching information about CVE ``"+args[0]+"``.")
	}
	return NewEmbed().
		SetTitle(args[0]).
		SetDescription(cveData.Summary).
		SetColor(0xC93130).MessageEmbed
}
