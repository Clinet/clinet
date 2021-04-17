package main

import (
	"errors"

	"github.com/JoshuaDoes/go-wolfram"
	"github.com/bwmarrin/discordgo"
)

// QueryService is the interface that a queryable service must satisfy
type QueryService interface {
	GetName() string
	GetColor() int
	GetIconURL() string
	Query(query string, env *QueryEnvironment) (*discordgo.MessageEmbed, error)
}

// QueryEnvironment holds the details about a query
type QueryEnvironment struct {
	Channel *discordgo.Channel //The channel the command was executed in
	Guild   *discordgo.Guild   //The guild the command was executed in
	Message *discordgo.Message //The message that triggered the command execution
	User    *discordgo.User    //The user that executed the command
	Member  *discordgo.Member  //The guild member that executed the command

	WolframConversation *wolfram.Conversation //The last recorded Wolfram|Alpha conversation

	//For use with custom response commands
	Command             string //The command used to execute the command with this environment (in the event of a command alias)
	BotPrefix           string //The bot prefix used to execute this command (useful for command lists and example commands)
	UpdatedMessageEvent bool
}

func initQueryServices() {
	botData.QueryServices = make([]QueryService, 0)

	if botData.BotOptions.UseCustomResponses {
		botData.QueryServices = append(botData.QueryServices, &QueryServiceCustomResponse{})
	}
	if gcpAuthTokenFile != "" {
		botData.QueryServices = append(botData.QueryServices, &QueryServiceGoogleAssistant{})
	}
	if botData.BotOptions.UseDuckDuckGo {
		botData.QueryServices = append(botData.QueryServices, &QueryServiceDuckDuckGo{})
	}
	if botData.BotOptions.UseWolframAlpha {
		botData.QueryServices = append(botData.QueryServices, &QueryServiceWolframAlpha{})
	}
}

func getQueryResult(query string, env *QueryEnvironment) (*discordgo.MessageEmbed, error) {
	for _, service := range botData.QueryServices {
		queryResult, err := service.Query(query, env)
		if err != nil {
			Error.Println(err)
			continue
		}

		queryResultEmbed := Embed{queryResult}

		if queryResult.Color == 0 {
			queryResultEmbed.SetColor(service.GetColor())
		}
		if service.GetIconURL() != "" {
			queryResultEmbed.SetFooter("Results from "+service.GetName()+".", service.GetIconURL())
		} else {
			queryResultEmbed.SetFooter("Results from " + service.GetName() + ".")
		}

		return queryResultEmbed.MessageEmbed, nil
	}

	return nil, errors.New("error finding service to handle url")
}
