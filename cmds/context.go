package cmds

import (
	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/json"
)

type CmdCtx struct {
	Alias          string                       //Alias that triggered command
	Args           []*CmdArg                    //Arguments for command handler
	Edited         bool                         //True when called in response to edited call
	User           *services.User               //User who called the command
	Channel        *services.Channel            //Channel where command was called
	Server         *services.Server             //Server where command was called
	Message        *services.Message            //Message that called this command
	Service        services.Service  `json:"-"` //Service client for service callbacks
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
func (ctx *CmdCtx) SetQuery(alias string) *CmdCtx {
	ctx.Alias  = alias  //helloworld
	return ctx
}
func (ctx *CmdCtx) SetEdited() *CmdCtx {
	ctx.Edited = true
	return ctx
}
func (ctx *CmdCtx) SetUser(user *services.User) *CmdCtx {
	ctx.User = user
	return ctx
}
func (ctx *CmdCtx) SetChannel(channel *services.Channel) *CmdCtx {
	ctx.Channel = channel
	return ctx
}
func (ctx *CmdCtx) SetServer(server *services.Server) *CmdCtx {
	ctx.Server = server
	return ctx
}
func (ctx *CmdCtx) SetMessage(message *services.Message) *CmdCtx {
	ctx.Message = message
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
func (ctx *CmdCtx) GetArg(name string) *CmdArg {
	for _, arg := range ctx.Args {
		if arg.Name == name {
			return arg
		}
	}
	return nil
}