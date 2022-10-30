package guilded

import (
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/convos"
	"github.com/Clinet/clinet/services"
)

func convoHandler(message *services.Message, session *ClientGuilded) (cmdResps []*cmds.CmdResp, err error) {
	if message == nil {
		return nil, cmds.ErrCmdEmptyMsg
	}
	content := message.Content
	if content == "" {
		return nil, nil
	}

	cmdResps = make([]*cmds.CmdResp, 0)

	conversation := convos.NewConversation()
	conversationState := conversation.QueryText(content)
	if len(conversationState.Errors) > 0 {
		for _, csErr := range conversationState.Errors {
			Log.Error(csErr)
		}
	}
	if conversationState.Response != nil {
		//TODO: Dynamically build either an embed response or a simple conversation response
		cmdResps = append(cmdResps, cmds.NewCmdRespMsg(conversationState.Response.TextSimple))
	} else {
		//TODO: Make a nice error message for failed queries in a conversation
		cmdResps = append(cmdResps, cmds.NewCmdRespMsg("I'm not sure how to respond to that yet!"))
	}

	return cmdResps, nil
}