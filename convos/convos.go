package convos

import (
	"time"
)

type Conversation struct {
	History []*ConversationState //Conversation state history in order
}

//NewConversation returns an empty conversation
func NewConversation() (*Conversation) {
	return &Conversation{
		History: make([]*ConversationState, 0),
	}
}

//QueryText returns a new conversation state for the given query text and appends it to the convo history
func (convo *Conversation) QueryText(queryText string) (*ConversationState) {
	newState := &ConversationState{
		Query: &ConversationQuery{
			Time: time.Now(),
			Text: queryText,
		},
		Errors: make([]error, 0),
	}

	/*//Query Google Assistant if no response yet
	if newState.Response == nil && GoogleAssistant != nil {
		resp, err := GoogleAssistant.Query(newState.Query, convo.LastState())
		if err != nil {
			newState.Errors = append(newState.Errors, err)
		} else {
			newState.Response = resp
		}
	}

	//Query DuckDuckGo if no response yet
	if newState.Response == nil && DuckDuckGo != nil {
		resp, err := DuckDuckGo.Query(newState.Query, nil)
		if err != nil {
			newState.Errors = append(newState.Errors, err)
		} else {
			newState.Response = resp
		}
	}*/

	//Query Wolfram|Alpha if no response yet
	if newState.Response == nil && WolframAlpha != nil {
		resp, err := WolframAlpha.Query(newState.Query, convo.LastState())
		if err != nil {
			newState.Errors = append(newState.Errors, err)
		} else {
			newState.Response = resp
		}
	}

	if newState.Response != nil {
		convo.History = append(convo.History, newState) //Only add successful responses to the history
	}

	return newState
}

//LastResponse returns the most recent conversation state
func (convo *Conversation) LastState() (*ConversationState) {
	if len(convo.History) == 0 {
		return nil
	}
	return convo.History[len(convo.History)-1]
}