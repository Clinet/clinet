package cmds

import (
	//"image/color"

	"github.com/JoshuaDoes/json"
)

type CmdResp struct {
	Ready  bool //When not ready, a typing event should be sent and a goroutine should wait on a response
	//Color  color.Color
	Color  *int
	Title  string
	Text   string
	Image  string
	Errors []error
}
func NewCmdRespMsg(text string) *CmdResp {
	return &CmdResp{Ready: true, Text: text}
}
func NewCmdRespEmbed(title, text string) *CmdResp {
	return &CmdResp{Ready: true, Title: title, Text: text}
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
//func (resp *CmdResp) SetColor(clr color.Color) *CmdResp {
func (resp *CmdResp) SetColor(clr int) *CmdResp {
	resp.Color = &clr
	return resp
}
func (resp *CmdResp) SetText(text string) *CmdResp {
	resp.Text = text
	return resp
}
func (resp *CmdResp) SetTitle(title string) *CmdResp {
	resp.Title = title
	return resp
}
func (resp *CmdResp) SetImage(image string) *CmdResp {
	resp.Image = image
	return resp
}
func (resp *CmdResp) AddError(err error) *CmdResp {
	resp.Errors = append(resp.Errors, err)
	return resp
}
