package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// GuildSettings holds settings specific to a guild
type GuildSettings struct { //By default this will only be configurable for users in a role with the server admin permission
	AllowVoice              bool                  `json:"allowVoice"`              //Whether voice commands should be usable in this guild
	BotAdminRoles           []string              `json:"adminRoles"`              //An array of role IDs that can admin the bot
	BotAdminUsers           []string              `json:"adminUsers"`              //An array of user IDs that can admin the bot
	BotOptions              BotOptions            `json:"botOptions"`              //The bot options to use in this guild (true gets overridden if global bot config is false)
	BotPrefix               string                `json:"botPrefix"`               //The bot prefix to use in this guild
	CustomResponses         []CustomResponseQuery `json:"customResponses"`         //An array of custom responses specific to the guild
	UserJoinMessage         string                `json:"userJoinMessage"`         //A message to send when a user joins
	UserJoinMessageChannel  string                `json:"userJoinMessageChannel"`  //The channel to send the user join message to
	UserLeaveMessage        string                `json:"userLeaveMessage"`        //A message to send when a user leaves
	UserLeaveMessageChannel string                `json:"userLeaveMessageChannel"` //The channel to send the user leave message to
}

// UserSettings holds settings specific to a user
type UserSettings struct {
	Balance     int    `json:"balance"`     //A balance to use as virtual currency for some bot tasks
	Description string `json:"description"` //A description set by the user
}

func discordUserJoin(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	_, guildFound := guildSettings[member.GuildID]
	if guildFound {
		debugLog("guild found", true)

		if guildSettings[member.GuildID].UserJoinMessage != "" && guildSettings[member.GuildID].UserJoinMessageChannel != "" {
			debugLog("channel "+guildSettings[member.GuildID].UserJoinMessageChannel+", message "+guildSettings[member.GuildID].UserJoinMessage, true)

			message := guildSettings[member.GuildID].UserJoinMessage
			message = strings.Replace(message, "{user}", member.User.Username, -1)
			message = strings.Replace(message, "{user-mention}", "<@"+member.User.ID+">", -1)
			message = strings.Replace(message, "{user-id}", member.User.ID, -1)
			message = strings.Replace(message, "{user-discriminator}", member.User.Discriminator, -1)

			session.ChannelMessageSend(guildSettings[member.GuildID].UserJoinMessageChannel, message)

			debugLog("channel "+guildSettings[member.GuildID].UserJoinMessageChannel+", message "+message, true)
		}
	} else {
		debugLog("guild not found", true)
	}
}
func discordUserLeave(session *discordgo.Session, member *discordgo.GuildMemberRemove) {
	_, guildFound := guildSettings[member.GuildID]
	if guildFound {
		if guildSettings[member.GuildID].UserLeaveMessage != "" && guildSettings[member.GuildID].UserLeaveMessageChannel != "" {
			message := guildSettings[member.GuildID].UserLeaveMessage
			message = strings.Replace(message, "{user}", member.User.Username, -1)
			message = strings.Replace(message, "{user-id}", member.User.ID, -1)
			message = strings.Replace(message, "{user-discriminator}", member.User.Discriminator, -1)

			session.ChannelMessageSend(guildSettings[member.GuildID].UserLeaveMessageChannel, message)
		}
	}
}
