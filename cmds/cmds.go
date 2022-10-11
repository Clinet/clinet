package cmds

import (
	"errors"
	"regexp"
)

var (
	ErrCmdEmptyMsg = errors.New("cmd: nil message")
	ErrCmdNotFound = errors.New("cmd: no commands to handle message")
	ErrCmdNoResp   = errors.New("cmd: no resp")
)

//Commands holds the complete command list
var Commands []*Cmd = make([]*Cmd, 0)

//GetCmd returns a cmd that matches the given alias
func GetCmd(match string) *Cmd {
	for _, cmd := range Commands {
		if match == cmd.Name {
			return cmd
		}
		for _, alias := range cmd.Aliases {
			if match == alias {
				return cmd
			}
		}
	}
	return nil
}

type Cmd struct {
	Exec        func(*CmdCtx) *CmdResp //Go function to handle command
	Name        string                 //Display name for command
	Description string                 //Description for command lists and command usage
	Aliases     []string               //Aliases to refer to command
	Regexes     []regexp.Regexp        //Regular expressions to match command call with natural language ($1 is Args[0], $2 is Args[1], etc)
	Args        []CmdArg               //Arguments for command
	Subcommands []*Cmd                 //Subcommands for command (nestable, i.e. "/minecraft server mc.hypixel.net" where server is subcommand to minecraft command)
}
func NewCmd(name, desc string, handler func(*CmdCtx) *CmdResp) *Cmd {
	return &Cmd{
		Name: name,
		Description: desc,
		Aliases: []string{name},
		Exec: handler,
	}
}
func (cmd *Cmd) SetHandler(handler func(*CmdCtx) *CmdResp) {
	cmd.Exec = handler
}
func (cmd *Cmd) SetName(name string) {
	cmd.Name = name
}
func (cmd *Cmd) SetDescription(desc string) {
	cmd.Description = desc
}
func (cmd *Cmd) AddAliases(alias ...string) {
	cmd.Aliases = append(cmd.Aliases, alias...)
}
func (cmd *Cmd) AddRegexes(regex ...regexp.Regexp) {
	cmd.Regexes = append(cmd.Regexes, regex...)
}
func (cmd *Cmd) AddArgs(arg ...CmdArg) {
	cmd.Args = append(cmd.Args, arg...)
}
func (cmd *Cmd) AddSubCmds(subCmd ...*Cmd) {
	cmd.Subcommands = append(cmd.Subcommands, subCmd...)
}
func (cmd *Cmd) IsAlias(alias string) bool {
	for i := 0; i < len(cmd.Aliases); i++ {
		if cmd.Aliases[i] == alias {
			return true
		}
	}
	return false
}