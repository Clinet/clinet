package cmds

import (
	"github.com/Clinet/clinet/services"
)

var (
	ArgTypeUser = &services.User{}
)

type CmdArg struct {
	Name        string      //Display name for argument
	Description string      //Description for command usage
	Value       interface{} //Value for argument, set to default value (or zero value if required argument) when creating command
	Required    bool        //True when argument must be changed
}
func NewCmdArg(name, desc string, value interface{}) *CmdArg {
	return &CmdArg{
		Name: name,
		Description: desc,
		Value: value,
	}
}
func (arg *CmdArg) GetInt() int {
	return arg.Value.(int)
}
func (arg *CmdArg) GetInt64() int {
	return arg.Value.(int)
}
func (arg *CmdArg) GetString() string {
	return arg.Value.(string)
}
func (arg *CmdArg) GetUser() *services.User {
	switch arg.Value.(type) {
	case string:
		return &services.User{UserID: arg.Value.(string)}
	case *services.User:
		return arg.Value.(*services.User)
	}
	return nil
}
func (arg *CmdArg) SetRequired() *CmdArg {
	arg.Required = true
	return arg
}