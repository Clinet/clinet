package discord

var DiscordCfg *CfgDiscord

//Configuration for Discord sessions
type CfgDiscord struct {
	//Stuff for communication with Discord
	Token string `json:"token"`

	//Trust for Discord communication
	OwnerID string `json:"ownerID"` //The user ID of the bot owner on Discord
}