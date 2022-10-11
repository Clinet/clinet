package convos

import (
	"errors"
	"strings"

	"github.com/JoshuaDoes/go-wolfram"
)

var WolframAlpha *ClientWolframAlpha

type ClientWolframAlpha struct {
	Client *wolfram.Client
}

func AuthWolframAlpha(client *wolfram.Client) {
	WolframAlpha = &ClientWolframAlpha{
		Client: client,
	}
}

func (wa *ClientWolframAlpha) Query(query *ConversationQuery, lastState *ConversationState) (*ConversationResponse, error) {
	resp := &ConversationResponse{
		TextSimple: "",
		WolframAlpha: nil,
	}
	if lastState != nil {
		resp.WolframAlpha = lastState.Response.WolframAlpha
	}

	wolframConvo, err := WolframAlpha.Client.GetConversationalQuery(query.Text, wolfram.Metric, resp.WolframAlpha)
	if err != nil {
		return nil, err
	}

	if wolframConvo.ErrorMessage != "" {
		return nil, errors.New("wolfram: " + wolframConvo.ErrorMessage)
	}

	if !strings.HasSuffix(wolframConvo.Result, ".") {
		wolframConvo.Result += "."
	}

	resp.TextSimple = wolframConvo.Result
	resp.WolframAlpha = wolframConvo

	return resp, nil
}