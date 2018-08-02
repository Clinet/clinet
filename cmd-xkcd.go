package main

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func commandXKCD(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "latest":
		comic, err := botData.BotClients.XKCD.Latest()
		if err != nil {
			return NewErrorEmbed("XKCD Error", "There was an error fetching the latest XKCD comic.")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	case "random":
		comic, err := botData.BotClients.XKCD.Random()
		if err != nil {
			return NewErrorEmbed("XKCD Error", "There was an error fetching a random XKCD comic.")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	default:
		comicNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return NewErrorEmbed("XKCD Error", "``"+args[0]+"`` is not a valid number.")
		}

		comic, err := botData.BotClients.XKCD.Get(comicNumber)
		if err != nil {
			return NewErrorEmbed("XKCD Error", "There was an error fetching XKCD comic #"+args[0]+".")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + args[0]).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	}
}