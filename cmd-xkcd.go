package main

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func commandXKCD(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "latest":
		latestComic, err := botData.BotClients.XKCD.Latest(context.Background())
		if err != nil {
			return NewErrorEmbed("xkcd Error", "There was an error fetching the latest xkcd comic.")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + strconv.Itoa(latestComic.Number)).
			SetDescription(latestComic.Title).
			SetImage(latestComic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	case "random":
		latestComic, err := botData.BotClients.XKCD.Latest(context.Background())
		if err != nil {
			return NewErrorEmbed("xkcd Error", "There was an error figuring out the latest xkcd comic number.")
		}

		randomComic, err := botData.BotClients.XKCD.Get(context.Background(), randomInRange(1, latestComic.Number+1))
		if err != nil {
			return NewErrorEmbed("xkcd Error", "There was an error fetching a random xkcd comic.")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + strconv.Itoa(randomComic.Number)).
			SetDescription(randomComic.Title).
			SetImage(randomComic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	default:
		comicNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return NewErrorEmbed("xkcd Error", "``"+args[0]+"`` is not a valid number.")
		}

		comic, err := botData.BotClients.XKCD.Get(context.Background(), comicNumber)
		if err != nil {
			return NewErrorEmbed("xkcd Error", "There was an error fetching xkcd comic #"+args[0]+".")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + args[0]).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	}
}
