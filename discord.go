package main

import (
	//Discord-related essentials
	discord "github.com/bwmarrin/discordgo" //Used to communicate with Discord
)

//Configuration for Discord sessions
type CfgDiscord struct {
	//Stuff for communication with Discord
	Token string `json:"token"`

	//Trust for Discord communication
	DisplayName   string `json:"displayName"`   //The display name for communicating on Discord
	OwnerID       string `json:"ownerID"`       //The user ID of the bot owner on Discord
	CommandPrefix string `json:"commandPrefix"` //The command prefix to use when invoking the bot on Discord
}

var Discord *discord.Session

func startDiscord() {
	log.Trace("--- startDiscord() ---")

	log.Debug("Creating Discord struct")
	Discord, err = discord.New("Bot " + cfg.Discord.Token)
	if err != nil {
		log.Fatal("Unable to connect to Discord!")
	}

	//Only enable informational Discord logging if we're tracing
	if log.Verbosity == 2 {
		log.Debug("Setting Discord log level to informational")
		Discord.LogLevel = discord.LogInformational
	}

	log.Info("Registering Discord event handlers")
	Discord.AddHandler(discordReady)
	Discord.AddHandler(discordChannelCreate)
	Discord.AddHandler(discordChannelUpdate)
	Discord.AddHandler(discordChannelDelete)
	Discord.AddHandler(discordGuildUpdate)
	Discord.AddHandler(discordGuildBanAdd)
	Discord.AddHandler(discordGuildBanRemove)
	Discord.AddHandler(discordGuildMemberAdd)
	Discord.AddHandler(discordGuildMemberRemove)
	Discord.AddHandler(discordGuildRoleCreate)
	Discord.AddHandler(discordGuildRoleUpdate)
	Discord.AddHandler(discordGuildRoleDelete)
	Discord.AddHandler(discordGuildEmojisUpdate)
	Discord.AddHandler(discordUserUpdate)
	Discord.AddHandler(discordVoiceStateUpdate)
	Discord.AddHandler(discordMessageCreate)
	Discord.AddHandler(discordMessageDelete)
	Discord.AddHandler(discordMessageDeleteBulk)
	Discord.AddHandler(discordMessageUpdate)
	Discord.AddHandler(discordMessageReactionAdd)
	Discord.AddHandler(discordMessageReactionRemove)
	Discord.AddHandler(discordMessageReactionRemoveAll)

	log.Info("Connecting to Discord")
	err = Discord.Open()
	if err != nil {
		log.Fatal("Unable to connect to Discord!", err)
	}
	log.Info("Connected to Discord!")
}

func closeDiscord() {
	log.Trace("--- closeDiscord() ---")

	log.Info("Closing connection to Discord...")
	Discord.Close()
}

func discordReady(session *discord.Session, event *discord.Ready) {
	log.Trace("--- discordReady(", event, ") ---")
}

func discordChannelCreate(session *discord.Session, event *discord.ChannelCreate) {
	log.Trace("--- discordChannelCreate(", event, ") ---")
}

func discordChannelUpdate(session *discord.Session, event *discord.ChannelUpdate) {
	log.Trace("--- discordChannelUpdate(", event, ") ---")
}

func discordChannelDelete(session *discord.Session, event *discord.ChannelDelete) {
	log.Trace("--- discordChannelDelete(", event, ") ---")
}

func discordGuildUpdate(session *discord.Session, event *discord.GuildUpdate) {
	log.Trace("--- discordGuildUpdate(", event, ") ---")
}

func discordGuildBanAdd(session *discord.Session, event *discord.GuildBanAdd) {
	log.Trace("--- discordGuildBanAdd(", event, ") ---")
}

func discordGuildBanRemove(session *discord.Session, event *discord.GuildBanRemove) {
	log.Trace("--- discordGuildBanRemove(", event, ") ---")
}

func discordGuildMemberAdd(session *discord.Session, event *discord.GuildMemberAdd) {
	log.Trace("--- discordGuildMemberAdd(", event, ") ---")
}

func discordGuildMemberRemove(session *discord.Session, event *discord.GuildMemberRemove) {
	log.Trace("--- discordGuildMemberRemove(", event, ") ---")
}

func discordGuildRoleCreate(session *discord.Session, event *discord.GuildRoleCreate) {
	log.Trace("--- discordGuildRoleCreate(", event, ") ---")
}

func discordGuildRoleUpdate(session *discord.Session, event *discord.GuildRoleUpdate) {
	log.Trace("--- discordGuildRoleUpdate(", event, ") ---")
}

func discordGuildRoleDelete(session *discord.Session, event *discord.GuildRoleDelete) {
	log.Trace("--- discordGuildRoleDelete(", event, ") ---")
}

func discordGuildEmojisUpdate(session *discord.Session, event *discord.GuildEmojisUpdate) {
	log.Trace("--- discordGuildEmojisUpdate(", event, ") ---")
}

func discordUserUpdate(session *discord.Session, event *discord.UserUpdate) {
	log.Trace("--- discordUserUpdate(", event, ") ---")
}

func discordVoiceStateUpdate(session *discord.Session, event *discord.VoiceStateUpdate) {
	log.Trace("--- discordVoiceStateUpdate(", event, ") ---")
}

func discordMessageCreate(session *discord.Session, event *discord.MessageCreate) {
	log.Trace("--- discordMessageCreate(", event, ") ---")
}

func discordMessageDelete(session *discord.Session, event *discord.MessageDelete) {
	log.Trace("--- discordMessageDelete(", event, ") ---")
}

func discordMessageDeleteBulk(session *discord.Session, event *discord.MessageDeleteBulk) {
	log.Trace("--- discordMessageDeleteBulkage(", event, ") ---")
}

func discordMessageUpdate(session *discord.Session, event *discord.MessageUpdate) {
	log.Trace("--- discordMessageUpdate(", event, ") ---")
}

func discordMessageReactionAdd(session *discord.Session, event *discord.MessageReactionAdd) {
	log.Trace("--- discordMessageReactionAdd(", event, ") ---")
}

func discordMessageReactionRemove(session *discord.Session, event *discord.MessageReactionRemove) {
	log.Trace("--- discordMessageReactionRemove(", event, ") ---")
}

func discordMessageReactionRemoveAll(session *discord.Session, event *discord.MessageReactionRemoveAll) {
	log.Trace("--- discordMessageReactionRemoveAll(", event, ") ---")
}
