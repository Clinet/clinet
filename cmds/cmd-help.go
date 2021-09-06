package cmds

func init() {
	Commands = append(Commands, &Cmd{
		Handler: cmdHelp,
		Matches: []string{"?", "help", "h", "about", "cmd", "cmds", "command", "commands"},
		Description: "Displays pages of available commands",
	})
}

func cmdHelp(ctx *CmdCtx) *CmdResp {
	help := `**cli$help** - Displays pages of available commands
**cli$hellodolly** - Returns a lyric from Hello Dolly`
	
	return makeCmdResp(help)
}