package main

import (
	"errors"
	"time"

	"github.com/bwmarrin/discordgo"
)

//GoogleAssistant holds a Google Assistant query service
type GoogleAssistant struct {
}

//GetName returns "Google Assistant"
func (*GoogleAssistant) GetName() string {
	return "Google Assistant"
}

//GetColor returns 0x0F9D58
func (*GoogleAssistant) GetColor() int {
	return 0x0F9D58
}

//GetIconURL returns "https://files.joshuadoes.com/googleassistant_transparent.png"
func (*GoogleAssistant) GetIconURL() string {
	return "https://files.joshuadoes.com/googleassistant_transparent.png"
}

//Query returns a Google Assistant response to a given query
func (*GoogleAssistant) Query(query string, env *QueryEnvironment) (*discordgo.MessageEmbed, error) {
	Debug.Println("[Google Assistant] Spawning a conversation...")
	conversation, err := botData.BotClients.GoogleAssistant.NewConversation(time.Second * 240)
	if err != nil {
		return nil, err
	}
	defer conversation.Close()

	Debug.Println("[Google Assistant] Requesting text transport...")
	textTransport := conversation.RequestTransportText()

	Debug.Printf("[Google Assistant] Getting result for [%s]...", query)
	result, err := textTransport.Query(query)
	if err != nil {
		return nil, err
	}

	if result == "" {
		Debug.Println("[Google Assistant] Error getting allowed result from response")
		return nil, errors.New("error getting allowed result from response")
	}

	gaEmbed := NewEmbed().
		AddField(query, result).MessageEmbed

	return gaEmbed, nil
}
