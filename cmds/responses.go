package cmds

import (
	//"image/color"

	"github.com/Clinet/clinet/services"
	"github.com/JoshuaDoes/json"
)

type CmdResp struct {
	*services.Message //Take on the fields of a service message
	Ready  bool       //When not ready, a typing event should be sent and a goroutine should wait on a response
}
func NewCmdRespMsg(content string) *CmdResp {
	resp := &CmdResp{&services.Message{Content: content}, true}
	return resp
}
func NewCmdRespEmbed(title, content string) *CmdResp {
	resp := &CmdResp{&services.Message{Title: title, Content: content}, true}
	return resp
}
func CmdRespFromMsg(msg *services.Message) *CmdResp {
	return &CmdResp{msg, false}
}
func (resp *CmdResp) String() string {
	jsonData, err := json.Marshal(resp, true)
	if err != nil {
		return err.Error()
	}
	return string(jsonData)
}
func (resp *CmdResp) OnReady(readyCall func(*CmdResp)) {
	go func(r *CmdResp) {
		for {
			if r == nil {
				return
			}
			if r.Ready {
				readyCall(r)
				return
			}
		}
	}(resp)
}
func (resp *CmdResp) SetReady(ready bool) *CmdResp {
	resp.Ready = ready
	return resp
}