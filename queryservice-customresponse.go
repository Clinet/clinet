package main

import (
	"errors"
	"math/rand"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

//QueryServiceCustomResponse exports the methods required to respond to custom queries
type QueryServiceCustomResponse struct {
}

//GetName returns "custom responses"
func (*QueryServiceCustomResponse) GetName() string {
	return "custom responses"
}

//GetColor returns 0x1C1C1C
func (*QueryServiceCustomResponse) GetColor() int {
	return 0x1C1C1C
}

//GetIconURL returns nothing
func (*QueryServiceCustomResponse) GetIconURL() string {
	return ""
}

//Query returns the response to a query
func (*QueryServiceCustomResponse) Query(query string, env *QueryEnvironment) (*discordgo.MessageEmbed, error) {
	//Initialize a local list of custom responses to be checked in order
	customResponses := make([]CustomResponseQuery, 0)

	//Add guild-specific custom responses
	if len(guildSettings[env.Guild.ID].CustomResponses) > 0 {
		customResponses = append(customResponses, guildSettings[env.Guild.ID].CustomResponses...)
	}
	//Add global custom responses
	if len(botData.CustomResponses) > 0 {
		customResponses = append(customResponses, botData.CustomResponses...)
	}

	for _, response := range customResponses {
		regexpMatched, _ := regexp.MatchString(response.Expression, query)
		if regexpMatched {
			if len(response.CmdResponses) > 0 {
				randomCmd := rand.Intn(len(response.CmdResponses))

				//Hotfix for missing command prefix in custom command responses
				prefix := guildSettings[env.Guild.ID].BotPrefix
				if prefix == "" {
					prefix = botData.CommandPrefix
				}

				commandEnvironment := &CommandEnvironment{
					Channel: env.Channel,
					Guild: env.Guild,
					Message: env.Message,
					User: env.User,
					Member: env.Member,
					Command: response.CmdResponses[randomCmd].CommandName,
					UpdatedMessageEvent: env.UpdatedMessageEvent,
					BotPrefix: prefix,
				}
				return callCommand(response.CmdResponses[randomCmd].CommandName, response.CmdResponses[randomCmd].Arguments, commandEnvironment), nil
			}
			if len(response.Responses) > 0 {
				random := rand.Intn(len(response.Responses))

				return response.Responses[random].ResponseEmbed, nil
			}
		}
	}

	return nil, errors.New("error finding custom response")
}
