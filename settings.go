package main

import (
	"strconv"
	"strings"
	"time"

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
	LogChannel              string                `json:"logChannel"`              //The channel to log guild events to
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

func discordChannelCreate(session *discordgo.Session, channel *discordgo.ChannelCreate) {

}
func discordChannelUpdate(session *discordgo.Session, channel *discordgo.ChannelUpdate) {

}
func discordChannelDelete(session *discordgo.Session, channel *discordgo.ChannelDelete) {

}
func discordGuildUpdate(session *discordgo.Session, guild *discordgo.GuildUpdate) {

}
func discordGuildBanAdd(session *discordgo.Session, guild *discordgo.GuildBanAdd) {

}
func discordGuildBanRemove(session *discordgo.Session, guild *discordgo.GuildBanRemove) {

}
func discordGuildMemberAdd(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	_, guildFound := guildSettings[member.GuildID]
	if guildFound {
		debugLog("guild found", true)

		if guildSettings[member.GuildID].UserJoinMessage != "" && guildSettings[member.GuildID].UserJoinMessageChannel != "" {
			message := guildSettings[member.GuildID].UserJoinMessage
			message = strings.Replace(message, "{user}", member.User.Username, -1)
			message = strings.Replace(message, "{user-mention}", "<@"+member.User.ID+">", -1)
			message = strings.Replace(message, "{user-id}", member.User.ID, -1)
			message = strings.Replace(message, "{user-discriminator}", member.User.Discriminator, -1)

			session.ChannelMessageSend(guildSettings[member.GuildID].UserJoinMessageChannel, message)
		}

		if guildSettings[member.GuildID].LogChannel != "" {
			joinedAt := member.JoinedAt
			joinedAtTimeFormatted := ""
			joinedAtTime, err := time.Parse(time.RFC3339Nano, joinedAt)
			if err != nil {
				joinedAtTimeFormatted = joinedAt
			} else {
				joinedAtMonth := joinedAtTime.Month().String()
				joinedAtDay := joinedAtTime.Day()
				joinedAtYear := joinedAtTime.Year()
				joinedAtHour := joinedAtTime.Hour()
				joinedAtMinute := joinedAtTime.Minute()
				joinedAtSecond := joinedAtTime.Second()
				joinedAtTimeFormatted = joinedAtMonth + " " + strconv.Itoa(joinedAtDay) + ", " + strconv.Itoa(joinedAtYear) + " at " + strconv.Itoa(joinedAtHour) + ":" + strconv.Itoa(joinedAtMinute) + ":" + strconv.Itoa(joinedAtSecond)
			}

			session.ChannelMessageSendEmbed(guildSettings[member.GuildID].LogChannel, NewEmbed().
				SetTitle("Server Log - User Joined").
				SetDescription("A new member joined the server.").
				AddField("Joined At", joinedAtTimeFormatted).
				AddField("User ID", member.User.ID).
				AddField("Username", member.User.Username).
				AddField("Verified Account", strconv.FormatBool(member.User.Verified)).
				AddField("Multi-Factor Authentication", strconv.FormatBool(member.User.MFAEnabled)).
				AddField("Bot", strconv.FormatBool(member.User.Bot)).
				SetImage(member.User.AvatarURL("2048")).
				SetColor(0x1C1C1C).MessageEmbed,
			)
		}
	} else {
		debugLog("guild not found", true)
	}
}
func discordGuildMemberRemove(session *discordgo.Session, member *discordgo.GuildMemberRemove) {
	_, guildFound := guildSettings[member.GuildID]
	if guildFound {
		if guildSettings[member.GuildID].UserLeaveMessage != "" && guildSettings[member.GuildID].UserLeaveMessageChannel != "" {
			message := guildSettings[member.GuildID].UserLeaveMessage
			message = strings.Replace(message, "{user}", member.User.Username, -1)
			message = strings.Replace(message, "{user-id}", member.User.ID, -1)
			message = strings.Replace(message, "{user-discriminator}", member.User.Discriminator, -1)

			session.ChannelMessageSend(guildSettings[member.GuildID].UserLeaveMessageChannel, message)
		}

		if guildSettings[member.GuildID].LogChannel != "" {
			joinedAt := member.JoinedAt
			joinedAtTimeFormatted := ""
			joinedAtTime, err := time.Parse(time.RFC3339Nano, joinedAt)
			if err != nil {
				joinedAtTimeFormatted = joinedAt
			} else {
				joinedAtMonth := joinedAtTime.Month().String()
				joinedAtDay := joinedAtTime.Day()
				joinedAtYear := joinedAtTime.Year()
				joinedAtHour := joinedAtTime.Hour()
				joinedAtMinute := joinedAtTime.Minute()
				joinedAtSecond := joinedAtTime.Second()
				joinedAtTimeFormatted = joinedAtMonth + " " + strconv.Itoa(joinedAtDay) + ", " + strconv.Itoa(joinedAtYear) + " at " + strconv.Itoa(joinedAtHour) + ":" + strconv.Itoa(joinedAtMinute) + ":" + strconv.Itoa(joinedAtSecond)
			}

			session.ChannelMessageSendEmbed(guildSettings[member.GuildID].LogChannel, NewEmbed().
				SetTitle("Server Log - User Left").
				SetDescription("A member left the server.").
				AddField("Joined At", joinedAtTimeFormatted).
				AddField("User ID", member.User.ID).
				AddField("Username", member.User.Username).
				AddField("Verified Account", strconv.FormatBool(member.User.Verified)).
				AddField("Multi-Factor Authentication", strconv.FormatBool(member.User.MFAEnabled)).
				AddField("Bot", strconv.FormatBool(member.User.Bot)).
				SetImage(member.User.AvatarURL("2048")).
				SetColor(0x1C1C1C).MessageEmbed,
			)
		}
	}
}
func discordGuildRoleCreate(session *discordgo.Session, guildRole *discordgo.GuildRoleCreate) {

}
func discordGuildRoleUpdate(session *discordgo.Session, guildRole *discordgo.GuildRoleUpdate) {

}
func discordGuildRoleDelete(session *discordgo.Session, guildRole *discordgo.GuildRoleDelete) {

}
func discordGuildEmojisUpdate(session *discordgo.Session, emojis *discordgo.GuildEmojisUpdate) {

}
func discordUserUpdate(session *discordgo.Session, user *discordgo.UserUpdate) {

}
func discordVoiceStateUpdate(session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate) {

}
