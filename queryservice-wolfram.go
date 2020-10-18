package main

import (
	"errors"
	"strings"

	"github.com/JoshuaDoes/go-wolfram"
	"github.com/bwmarrin/discordgo"
)

//QueryServiceWolframAlpha exports the methods required to use Wolfram Alpha
type QueryServiceWolframAlpha struct {
}

//GetName returns "Wolfram|Alpha"
func (*QueryServiceWolframAlpha) GetName() string {
	return "Wolfram|Alpha"
}

//GetColor returns 0xDA0E1A
func (*QueryServiceWolframAlpha) GetColor() int {
	return 0xDA0E1A
}

//GetIconURL returns "https://joshuadoes.com/WolframAlpha.png"
func (*QueryServiceWolframAlpha) GetIconURL() string {
	return "https://joshuadoes.com/WolframAlpha.png"
}

//Query returns the response to a query
func (service *QueryServiceWolframAlpha) Query(query string, env *QueryEnvironment) (*discordgo.MessageEmbed, error) {
	Debug.Printf("[Wolfram|Alpha] Getting result for query [%s]...", query)
	conversationResult, err := botData.BotClients.Wolfram.GetConversationalQuery(query, wolfram.Metric, env.WolframConversation)
	if err != nil {
		Debug.Printf("[Wolfram|Alpha] Error getting query result: %v", err)
		wolframStoreConversation(nil, env)
		return nil, errors.New("error getting spoken answer from Wolfram|Alpha")
	}

	if conversationResult.ErrorMessage != "" {
		Debug.Printf("[Wolfram|Alpha] Error getting query result: %s", conversationResult.ErrorMessage)
		if env.WolframConversation != nil {
			Debug.Printf("[Wolfram|Alpha] Attempting nil conversation for query [%s]...", query)
			env.WolframConversation = nil
			return service.Query(query, env)
		}
		return nil, errors.New(conversationResult.ErrorMessage)
	}

	wolframStoreConversation(conversationResult, env)

	if !strings.HasSuffix(conversationResult.Result, ".") {
		conversationResult.Result += "."
	}

	result := conversationResult.Result

	for old, new := range botData.BotOptions.QueryResponseReplacements {
		result = strings.ReplaceAll(result, old, new)
	}

	wolframEmbed := NewEmbed().
		AddField(query, result).MessageEmbed

	return wolframEmbed, nil
}

func wolframStoreConversation(conversation *wolfram.Conversation, env *QueryEnvironment) {
	Debug.Printf("[Wolfram|Alpha] Storing conversation...")
	guildData[env.Guild.ID].WolframConversations[env.User.ID] = conversation
}
