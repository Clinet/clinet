package cmds

import (
	"github.com/bwmarrin/discordgo"
)

//Cmd holds a command that can be executed with arguments
type Cmd struct {
	Handler     func(*CmdCtx) *CmdResp //Go handler for this command
	Matches     []string               //Aliases for this command
	Description string                 //Description of this command
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
	Message *discordgo.Message //Discord message
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

//Holds the complete command list
var Commands []*Cmd = []*Cmd{
	&Cmd{
		Handler: HelloDolly,
		Matches: []string{"hellodolly", "hd", "hidolly", "dolly"},
		Description: "This is not just a command, it symbolizes the hope and enthusiasm of an entire generation summed up in two words sung most famously by Louis Armstrong: Hello, Dolly. When executed, you will randomly receive a lyric from Hello, Dolly as your response.",
	},
	&Cmd{
		Handler: makeCmdRespHandler("Hello, world!"),
		Matches: []string{"helloworld", "hw"},
		Description: "Simply responds with a hello, world!",
	},
}
