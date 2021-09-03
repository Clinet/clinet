package main

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
)

var (
	errCmdEmptyMsg = errors.New("cmd: nil message")
	errCmdNotFound = errors.New("cmd: no commands to handle message")
)

//CmdBuilder builds a Cmd
type CmdBuilder struct {
	Commands []*CmdBuilderCommand
}
type CmdBuilderCommand struct {
	Command *cmds.Cmd    //Command to execute
	Context *cmds.CmdCtx //Context for command
}
func CmdBuild(session *discordgo.Session, message *discordgo.Message) (*CmdBuilder, error) {
	if message == nil {
		return nil, errCmdEmptyMsg
	}

	content := message.Content
	if content[:len(cfg.Discord.CommandPrefix)] != cfg.Discord.CommandPrefix {
		return nil, nil
	}
	content = content[len(cfg.Discord.CommandPrefix):]

	//Prepare cmd context
	contentDisplay, err := message.ContentWithMoreMentionsReplaced(session)
	if err != nil {
		return nil, err
	}
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		return nil, err
	}
	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		return nil, err
	}
	author := message.Author
	member, err := session.GuildMember(guild.ID, author.ID)
	if err != nil {
		return nil, err
	}

	//Build cmd
	var cmd *cmds.Cmd = nil
	alias := ""

	raw := strings.Split(content, " ")
	for i := 0; i < len(cmds.Commands); i++ {
		testMatches := cmds.Commands[i].Matches
		for j := 0; j < len(testMatches); j++ {
			if testMatches[j] == raw[0] {
				cmd = cmds.Commands[i]
				alias = raw[0]
				break
			}
		}
		if alias != "" {
			break
		}
	}

	if cmd == nil {
		return nil, errCmdNotFound
	}

	//Build cmd context
	ctx := &cmds.CmdCtx{
		Content: content,
		ContentDisplay: contentDisplay,
		CmdAlias: alias,
		CmdParams: make([]string, 0),
		CmdPrefix: "cli$",
		CmdEdited: false,
		Server: guild,
		Channel: channel,
		Message: message,
		User: author,
		Member: member,
	}

	//Build cmd builder
	cmdBuilder := &CmdBuilder{
		Commands: make([]*CmdBuilderCommand, 0),
	}
	cmdBuilder.Commands = append(cmdBuilder.Commands, &CmdBuilderCommand{
		Command: cmd,
		Context: ctx,
	})

	return cmdBuilder, nil
}
func (cmdBuild *CmdBuilder) Run() ([]*cmds.CmdResp) {
	if cmdBuild == nil || len(cmdBuild.Commands) == 0 {
		return nil
	}
	resps := make([]*cmds.CmdResp, 0)
	for _, command := range cmdBuild.Commands {
		resps = append(resps, command.Command.Exec(command.Context))
	}
	return resps
}
