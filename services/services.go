package services

import (
	"fmt"
)

//Service requires various methods for the rest of the command framework to function.
// A dummy service can be used if you're looking to import a particular feature absent a service.
type Service interface {
	//Messages are the backbone of how the command framework responds to interactions and interacts with the service.
	// For services only able to process text messages, you can use Message.String() to get a preformatted text when necessary.
	// In the case of other concepts such as Discord's interaction events, use Message.Context to track the alternative type.
	MsgEdit(msg *Message) (ret *Message, err error) //Edits any type of message
	MsgRemove(msg *Message) (err error)             //Removes a message
	MsgSend(msg *Message) (ret *Message, err error) //Sends any type of message
}

func Error(format string, replacements ...interface{}) error {
	return fmt.Errorf(format, replacements)
}

type Message struct {
	AuthorID  string          `json:"authorID,omitempty"`
	MessageID string          `json:"messageID,omitempty"`
	ChannelID string          `json:"channelID,omitempty"`
	ServerID  string          `json:"serverID,omitempty"`
	Title     string          `json:"title,omitempty"`
	Content   string          `json:"content,omitempty"`
	Image     string          `json:"image,omitempty"`
	Color     *int            `json:"color,omitempty"`
	Fields    []*MessageField `json:"fields,omitempty"`
	Context   interface{}     `json:"context,omitempty"`
}

type MessageField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}