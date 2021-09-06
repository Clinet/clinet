package cmds

import (
	"errors"
	"strings"
)

var (
	ErrCmdEmptyMsg = errors.New("cmd: nil message")
	ErrCmdNotFound = errors.New("cmd: no commands to handle message")
	ErrCmdNoResp   = errors.New("cmd: no resp")
)

//Commands holds the complete command list
var Commands []*Cmd = []*Cmd{
	&Cmd{
		Handler: makeCmdRespHandler("Hello, world!"),
		Matches: []string{"helloworld", "hw"},
		Description: "Simply responds with a hello, world!",
	},
}

//GetCmd returns a cmd that matches the given alias
func GetCmd(alias string) *Cmd {
	for _, cmd := range Commands {
		for _, match := range cmd.Matches {
			if match == alias {
				return cmd
			}
		}
	}
	return nil
}

//CmdBuilder builds a list of Cmd paired to a CmdCtx
type CmdBuilder struct {
	Commands []*CmdBuilderCommand
}
func CmdBatch(cmds ...*CmdBuilderCommand) *CmdBuilder {
	return &CmdBuilder{
		Commands: cmds,
	}
}
type CmdBuilderCommand struct {
	Command *Cmd    //Command to execute
	Context *CmdCtx //Context for command
}
func CmdBuildCommand(cmd *Cmd, ctx *CmdCtx) *CmdBuilderCommand {
	return &CmdBuilderCommand{Command: cmd, Context: ctx}
}
func (cmdBuild *CmdBuilder) Run() ([]*CmdResp) {
	if cmdBuild == nil || len(cmdBuild.Commands) == 0 {
		return nil
	}
	resps := make([]*CmdResp, 0)
	for _, command := range cmdBuild.Commands {
		resps = append(resps, command.Command.Exec(command.Context))
	}
	return resps
}

func CmdMessage(ctx *CmdCtx, content string) (*CmdBuilderCommand, error) {
	//Build cmd
	var cmd *Cmd = nil

	raw := strings.Split(content, " ")
	cmd = GetCmd(raw[0])
	if cmd == nil {
		return nil, ErrCmdNotFound
	}

	//Process ctx
	ctx.CmdAlias = raw[0]
	ctx.CmdParams = make([]string, 0)
	if len(raw) > 1 {
		ctx.CmdParams = raw[1:]
	}

	//Return cmd builder command
	return CmdBuildCommand(cmd, ctx), nil
}