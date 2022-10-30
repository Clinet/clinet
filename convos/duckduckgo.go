package convos

import (
	"errors"

	duckduckgo "github.com/JoshuaDoes/duckduckgolang"
)

var DuckDuckGo *ClientDuckDuckGo

type ClientDuckDuckGo struct {
	Client *duckduckgo.Client
}

func AuthDuckDuckGo(client *duckduckgo.Client) {
	DuckDuckGo = &ClientDuckDuckGo{
		Client: client,
	}
}

func (ddg *ClientDuckDuckGo) Query(query *ConversationQuery, lastState *ConversationState) (*ConversationResponse, error) {
	resp := &ConversationResponse{}

	queryResult, err := ddg.Client.GetQueryResult(query.Text)
	if err != nil {
		return nil, err
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
		return nil, errors.New("duckduckgo: empty result")
	}
	resp.TextSimple = result

	if queryResult.Image != "" {
		resp.ImageURL = "https://duckduckgo.com" + queryResult.Image
	}

	return resp, nil
}