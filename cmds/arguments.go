package cmds

import (
	"fmt"
	"strconv"

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
	switch arg.Value.(type) {
	case string:
		num, err := strconv.Atoi(arg.Value.(string))
		if err != nil {
			return 0
		}
		return num
	case int64:
		return int(arg.Value.(int64))
	}
	return arg.Value.(int)
}
func (arg *CmdArg) GetInt64() int64 {
	switch arg.Value.(type) {
	case string:
		num, err := strconv.Atoi(arg.Value.(string))
		if err != nil {
			return 0
		}
		return int64(num)
	case int:
		return int64(arg.Value.(int))
	}
	return arg.Value.(int64)
}
func (arg *CmdArg) GetString() string {
	switch arg.Value.(type) {
	case int:
		return fmt.Sprintf("%d", arg.Value.(int))
	case int64:
		return fmt.Sprintf("%d", arg.Value.(int64))
	}
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