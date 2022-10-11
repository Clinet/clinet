package cmds

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