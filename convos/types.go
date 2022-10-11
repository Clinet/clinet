package convos

import (
	"time"

	"github.com/JoshuaDoes/go-wolfram"
)

type ConversationState struct {
	Query          *ConversationQuery    `json:"query"`    //Conversation query
	Response       *ConversationResponse `json:"response"` //Conversation response
	Errors         []error               `json:"errors"`   //Errors encountered while processing this state
}

type ConversationQuery struct {
	Time  time.Time `json:"time"`  //Time of query request
	Text  string     `json:"text"` //Query as transcribed
	Audio []byte     `json:"-"`    //Raw audio query in any ffmpeg-supported format
}

type ConversationResponse struct {
	TextSimple string      `json:"text"`   //Simple text response to query
	TextFields []TextField `json:"fields"` //Advanced text response to query
	Audio      []byte      `json:"-"`      //Raw audio response
	ImageURL   string      `json:"image"`  //URL to image supplied with response
	VideoURL   string      `json:"video"`  //URL to video supplied with response

	//Service conversation states
	ExpirationTime  *time.Time            `json:"expiresAt"`    //When to expire this response's service conversation states, nil to leave unchecked
	WolframAlpha    *wolfram.Conversation `json:"stateWolfram"` //Conversation state from Wolfram|Alpha, if present
	//GoogleAssistant *gassist.Conversation `json:"stateGoogle"`  //Conversation state from Google Assistant, if present
}

type TextField struct {
	Header  string `json:"header"`  //Field header/title
	Subtext string `json:"subtext"` //Field subtext/description
}
func NewTextField(header, subtext string) TextField {
	return TextField{Header: header, Subtext: subtext}
}