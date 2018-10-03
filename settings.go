package main

// GuildSettings holds settings specific to a guild
type GuildSettings struct { //By default this will only be configurable for users in a role with the server admin permission
	AllowVoice              bool                  `json:"allowVoice"`              //Whether voice commands should be usable in this guild
	BotAdminRoles           []string              `json:"adminRoles"`              //An array of role IDs that can admin the bot without the guild administrator permission
	BotAdminUsers           []string              `json:"adminUsers"`              //An array of user IDs that can admin the bot without a guild administrator role
	BotOptions              BotOptions            `json:"botOptions"`              //The bot options to use in this guild (true gets overridden if global bot config is false)
	BotPrefix               string                `json:"botPrefix"`               //The bot prefix to use in this guild
	CustomResponses         []CustomResponseQuery `json:"customResponses"`         //An array of custom responses specific to the guild
	LogSettings             LogSettings           `json:"logSettings"`             //Logging settings
	SwearFilter             SwearFilter           `json:"swearFilter"`             //The swear filter settings specific to this guild
	UserJoinMessage         string                `json:"userJoinMessage"`         //A message to send when a user joins
	UserJoinMessageChannel  string                `json:"userJoinMessageChannel"`  //The channel to send the user join message to
	UserLeaveMessage        string                `json:"userLeaveMessage"`        //A message to send when a user leaves
	UserLeaveMessageChannel string                `json:"userLeaveMessageChannel"` //The channel to send the user leave message to
}

// UserSettings holds settings specific to a user
type UserSettings struct {
	Balance     int    `json:"balance"`     //A balance to use as virtual currency for some bot tasks
	Description string `json:"description"` //A description set by the user
	Timezone    string `json:"timezone"`    //A timezone set by the user to use in other functions
}
