package cmds

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
func (arg *CmdArg) GetString() string {
	return arg.Value.(string)
}
func (arg *CmdArg) SetRequired() {
	arg.Required = true
}