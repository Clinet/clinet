package cmds

import (
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/json"
)

type CmdCtx struct {
	Content        string           //Raw message that triggered command
	ContentDisplay string           //Displayable form of raw message
	Alias          string           //Alias that triggered command
	Args           []*CmdArg        //Arguments for command handler
	Edited         bool             //True when called in response to edited call
	Service        services.Service //Service client for service callbacks
}
func NewCmdCtx() *CmdCtx {
	return &CmdCtx{}
}
func (ctx *CmdCtx) String() string {
	jsonData, err := json.Marshal(ctx, true)
	if err != nil {
		return err.Error()
	}
	return string(jsonData)
}
func (ctx *CmdCtx) SetAlias(alias string) *CmdCtx {
	ctx.Alias = alias
	return ctx
}
func (ctx *CmdCtx) SetContent(content, contentDisplay string) *CmdCtx {
	ctx.Content        = content        //<@!xxxxxxxxxx> Hello, world!
	ctx.ContentDisplay = contentDisplay //@Clinet Hello, world!
	return ctx
}
func (ctx *CmdCtx) SetQuery(alias string) *CmdCtx {
	ctx.Alias  = alias  //helloworld
	return ctx
}
func (ctx *CmdCtx) SetEdited() *CmdCtx {
	ctx.Edited = true
	return ctx
}
func (ctx *CmdCtx) SetService(service services.Service) *CmdCtx {
	ctx.Service = service
	return ctx
}
func (ctx *CmdCtx) AddArgs(arg ...*CmdArg) *CmdCtx {
	ctx.Args = append(ctx.Args, arg...)
	return ctx
}