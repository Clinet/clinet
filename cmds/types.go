package cmds

import (
	"github.com/bwmarrin/discordgo"
)

//Cmd holds a command that can be executed with arguments
type Cmd struct {
	Handler     func(*CmdCtx) *CmdResp //Go handler for this command
	Matches     []string               //Aliases for this command
	Description string                 //Description of this command
	Usage       []string               //How to use this command
}
func BuildCmd(handler func(*CmdCtx) *CmdResp, matches []string, description string) *Cmd {
	return &Cmd{Handler: handler, Matches: matches, Description: description}
}
func (cmd *Cmd) IsAlias(alias string) bool {
	for _, match := range cmd.Matches {
		if alias == match {
			return true
		}
	}
	return false
}
func (cmd *Cmd) Exec(ctx *CmdCtx) (*CmdResp) {
	return cmd.Handler(ctx)
}

type CmdCtx struct {
	Content        string   //The raw message that triggered this command
	ContentDisplay string   //The displayable message that triggered this command
	CmdAlias       string   //Alias used to run command
	CmdParams      []string //Input parameters
	CmdPrefix      string   //Prefix used to run command
	CmdEdited      bool     //Set when responding to an edited query

	//Discord environment
	Server  *discordgo.Guild   //Discord guild/server
	Channel *discordgo.Channel //Discord channel
	Message *discordgo.Message //Discord message that triggered this command
	User    *discordgo.User    //Discord message author
	Member  *discordgo.Member  //Discord guild member
}

type CmdResp struct {
	Messages []string
	Error    error
}
func makeCmdResp(messages ...string) *CmdResp {
	return &CmdResp{Messages: messages}
}
func makeCmdRespHandler(messages ...string) (func(*CmdCtx) *CmdResp) {
	return func(ctx *CmdCtx) *CmdResp {
		return makeCmdResp(messages...)
	}
}
