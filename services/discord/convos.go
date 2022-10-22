package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/convos"
)

func convoHandler(session *discordgo.Session, message *discordgo.Message) (cmdResps []*cmds.CmdResp, err error) {
	if message == nil {
		return nil, cmds.ErrCmdEmptyMsg
	}
	if message.Author.Bot {
		return nil, nil
	}
	content := message.Content
	if content == "" {
		return nil, nil
	}

	//Determine interaction type and how it was called
	prefix := ""
	convo := false
	if strings.Contains(content, "<@" + Discord.User.ID + ">") {
		prefix = "<@" + Discord.User.ID + ">"
		convo = true
	} else if strings.Contains(content, "<@!" + Discord.User.ID + ">") {
		prefix = "<@!" + Discord.User.ID + ">"
		convo = true
	}

	if !convo {
		return nil, nil
	}
	content = strings.ReplaceAll(content, prefix, "")

	cmdResps = make([]*cmds.CmdResp, 0)

	conversation := convos.NewConversation()
	conversationState := conversation.QueryText(content)
	if len(conversationState.Errors) > 0 {
		cmdResps = append(cmdResps, cmds.NewCmdRespMsg("No responses from known sources!"))
		return nil, nil
	}
	if conversationState.Response != nil {
		//TODO: Dynamically build either an embed response or a simple conversation response
		cmdResps = append(cmdResps, cmds.NewCmdRespMsg(conversationState.Response.TextSimple))
	} else {
		//TODO: Make a nice error message for failed queries in a conversation
		cmdResps = append(cmdResps, cmds.NewCmdRespMsg("I'm not sure how to respond to that yet!"))
	}

	//Return the responses
	if len(cmdResps) == 0 {
		Log.Error("no responses")
		return nil, cmds.ErrCmdNoResp
	}
	return cmdResps, nil
}