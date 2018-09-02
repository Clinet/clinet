package main

// LogSettings holds settings specific to logging
type LogSettings struct {
	LoggingEnabled bool      `json:"loggingEnabled"` //Whether or not logging enabled
	LoggingChannel string    `json:"loggingChannel"` //The channel to log guild events to
	LoggingEvents  LogEvents `json:"loggingEvents"`  //The events to log
}

// LogEvents holds logging events and whether or not they're enabled
type LogEvents struct {
	//Events received from Discord
	ChannelCreate     bool `json:"channelCreate"`
	ChannelUpdate     bool `json:"channelUpdate"`
	ChannelDelete     bool `json:"channelDelete"`
	GuildUpdate       bool `json:"guildUpdate"`
	GuildBanAdd       bool `json:"guildBanAdd"`
	GuildBanRemove    bool `json:"guildBanRemove"`
	GuildMemberAdd    bool `json:"guildMemberAdd"`
	GuildMemberRemove bool `json:"guildMemberRemove"`
	GuildRoleCreate   bool `json:"guildRoleCreate"`
	GuildRoleUpdate   bool `json:"guildRoleUpdate"`
	GuildRoleDelete   bool `json:"guildRoleDelete"`
	GuildEmojisUpdate bool `json:"guildEmojisUpdate"`
	UserUpdate        bool `json:"userUpdate"`
	VoiceStateUpdate  bool `json:"voiceStateUpdate"`

	//Custom events
	SwearDetect bool `json:"swearDetect"` //Triggered if a user uses a blacklisted (swear) word
	UserModlog  bool `json:"userModlog"`  //Triggered if a user's modlog is updated globally
}

var (
	// LogEventsRecommended contains pre-enabled recommended logging events
	LogEventsRecommended = LogEvents{
		ChannelCreate:     true,
		ChannelDelete:     true,
		GuildUpdate:       true,
		GuildBanAdd:       true,
		GuildBanRemove:    true,
		GuildMemberAdd:    true,
		GuildMemberRemove: true,
		GuildRoleCreate:   true,
		GuildRoleUpdate:   true,
		GuildRoleDelete:   true,
		VoiceStateUpdate:  true,
		SwearDetect:       true,
		UserModlog:        true,
	}
)
