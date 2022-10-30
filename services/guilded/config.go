package guilded

var GuildedCfg *CfgGuilded

//Configuration for Guilded sessions
type CfgGuilded struct {
	//Stuff for communication with Guilded
	BotName   string `json:"botName"`
	CmdPrefix string `json:"cmdPrefix"`
	Token     string `json:"token"`

	//Trust for Guilded communication
	OwnerID string `json:"ownerID"` //The user ID of the bot owner on Guilded
}