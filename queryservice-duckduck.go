package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

//QueryServiceDuckDuckGo exports the methods required to use DuckDuckGo
type QueryServiceDuckDuckGo struct {
}

//GetName returns "DuckDuckGo"
func (*QueryServiceDuckDuckGo) GetName() string {
	return "DuckDuckGo"
}

//GetColor returns 0xDF5730
func (*QueryServiceDuckDuckGo) GetColor() int {
	return 0xDF5730
}

//GetIconURL returns "https://upload.wikimedia.org/wikipedia/en/9/90/The_DuckDuckGo_Duck.png"
func (*QueryServiceDuckDuckGo) GetIconURL() string {
	return "https://upload.wikimedia.org/wikipedia/en/9/90/The_DuckDuckGo_Duck.png"
}

//Query returns the response to a query
func (*QueryServiceDuckDuckGo) Query(query string, env *QueryEnvironment) (*discordgo.MessageEmbed, error) {
	Debug.Printf("[DuckDuckGo] Getting result for [%s]...", query)
	queryResult, err := botData.BotClients.DuckDuckGo.GetQueryResult(query)
	if err != nil {
		Debug.Printf("[DuckDuckGo] Error getting query result: %v", err)
		return nil, errors.New("error getting response")
	}

	result := ""
	if queryResult.Definition != "" {
		result = queryResult.Definition
	} else if queryResult.Answer != "" {
		result = queryResult.Answer
	} else if queryResult.AbstractText != "" {
		result = queryResult.AbstractText
	}

	if result == "" {
		Debug.Println("[DuckDuckGo] Error getting allowed result from response")
		return nil, errors.New("error getting allowed result from response")
	}

	ddgEmbed := NewEmbed().
		AddField(queryResult.Heading, result).MessageEmbed

	if queryResult.Image != "" {
		ddgEmbed.Image = &discordgo.MessageEmbedImage{URL: queryResult.Image}
	}

	return ddgEmbed, nil
}
