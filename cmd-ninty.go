package main

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

func commandNNID(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	exists, exml, err := botData.BotClients.Ninty.DoesUserExist(args[0])
	if err != nil {
		return NewErrorEmbed("NNID Error", "Error checking for user ``"+args[0]+"``.")
	}

	if len(exml.Errors) != 0 {
		return NewErrorEmbed("NNID Error", "Error checking for user ``"+args[0]+"``.\n```"+exml.Errors[0].Error()+"```")
	}

	if exists {
		pids, exml, err := botData.BotClients.Ninty.GetPIDs(args)
		if err != nil {
			return NewErrorEmbed("NNID Error", "Error getting pid for user ``"+args[0]+"``.")
		}

		if len(exml.Errors) != 0 {
			return NewErrorEmbed("NNID Error", "Error getting pid for user ``"+args[0]+"``.\n```"+exml.Errors[0].Error()+"```")
		}

		miis, exml, err := botData.BotClients.Ninty.GetMiis(pids)
		if err != nil {
			return NewErrorEmbed("NNID Error", "Error getting mii for user ``"+args[0]+"``.")
		}

		if len(exml.Errors) != 0 {
			return NewErrorEmbed("NNID Error", "Error getting mii for user ``"+args[0]+"``.\n```"+exml.Errors[0].Error()+"```")
		}

		e := NewEmbed().
			SetTitle("NNID").
			SetDescription("Showing information for user ``" + args[0] + "``.").
			SetColor(0xf78b33)

		for i, miiImage := range miis.Miis[0].Images {
			if miiImage.Type == "normal_face" || i == len(miis.Miis[0].Images)-1 {
				e.SetImage(miiImage.URL, miiImage.CachedURL)
				break
			}
		}

		e.AddField("Name", miis.Miis[0].Name).
			AddField("PID", strconv.FormatInt(miis.Miis[0].PID, 10)).
			InlineAllFields()

		return e.MessageEmbed
	}

	return NewGenericEmbed("NNID", "The user ``"+args[0]+"`` does not exist.")
}
