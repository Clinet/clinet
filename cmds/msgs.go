package cmds

import (
	"strings"

	"github.com/Clinet/clinet/services"
)

func CmdHandler(serviceMsg *services.Message, service services.Service) (string, []*CmdResp, error) {
	if serviceMsg == nil || serviceMsg.Content == "" {
		return "", nil, services.Error("no message content to parse")
	}

	cmdPrefix := service.CmdPrefix()

	msgContent := strings.ReplaceAll(serviceMsg.Content, "\r\n", "\n")
	msgs := strings.Split(msgContent, "\n")

	//TODO: Loop msgs to create and handle msg
	rawMsg := msgs[0]
	if cmdPrefix != "" {
		if len(rawMsg) <= len(cmdPrefix) {
			return "", nil, nil
		}
		if string(rawMsg[:len(cmdPrefix)]) != cmdPrefix {
			return "", nil, nil
		}
		rawMsg = strings.Replace(rawMsg, cmdPrefix, "", 1) //Replace only the leftmost instance of the cmdPrefix
	}

	msg := strings.Split(rawMsg, " ")
	cmd := GetCmd(msg[0])
	if cmd == nil {
		return "", nil, nil
	}

	cmdArgs := cmd.Args
	isArgs := false
	if len(msg) > 1 {
		for i := 1; i < len(msg); i++ {
			kv := make([]string, 0)
			if strings.Contains(msg[i], ":") {
				isArgs = true
				kv = strings.Split(msg[i], ":")
				if len(kv) != 2 {
					return "", nil, services.Error("cmd arg must be key:value pair")
				}
			} else if strings.Contains(msg[i], "=") {
				isArgs = true
				kv = strings.Split(msg[i], "=")
				if len(kv) != 2 {
					return "", nil, services.Error("cmd arg must be key=value pair")
				}
			} else {
				if !isArgs {
					cmd = cmd.GetSubCmd(msg[i])
					if cmd == nil {
						return "", nil, services.Error("unknown subcommand %s", msg[i])
					}
				} else {
					return "", nil, services.Error("must specify value for %s with %s:value or %s=value", msg[i], msg[i], msg[i])
				}
			}

			if len(kv) != 0 {
				setValue := false
				for j := 0; j < len(cmdArgs); j++ {
					if cmdArgs[j].Name == kv[0] {
						cmdArgs[j].Value = kv[1]
						setValue = true
						break
					}
				}
				if !setValue {
					return "", nil, services.Error("cmd arg %s does not exist", kv[0])
				}
			}
		}
	}

	if len(cmd.Subcommands) > 0 {
		return "", nil, services.Error("must specify one of available subcommands")
	}

	for i := 0; i < len(cmdArgs); i++ {
		if cmdArgs[i].Required && cmdArgs[i].Value == nil {
			return "", nil, services.Error("must specify cmd arg %s", cmdArgs[i].Name)
		}
	}

	user := &services.User{
		ServerID: serviceMsg.ServerID,
		UserID: serviceMsg.UserID,
	}
	channel := &services.Channel{
		ServerID: serviceMsg.ServerID,
		ChannelID: serviceMsg.ChannelID,
	}
	server := &services.Server{
		ServerID: serviceMsg.ServerID,
	}

	cmdCtx := NewCmdCtx().
		SetAlias(cmd.Name).
		SetUser(user).
		SetChannel(channel).
		SetServer(server).
		SetMessage(serviceMsg).
		SetService(service).
		AddArgs(cmdArgs...)
	cmdBuilder := &CmdBuilderCommand{Command: cmd, Context: cmdCtx}
	cmdRuntime := CmdBatch(cmdBuilder)

	cmdResps := cmdRuntime.Run()
	return cmd.Name, cmdResps, nil
}