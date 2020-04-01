package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

/*
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
*/

func discordReady(session *discord.Session, event *discord.Ready) {
	Log.Trace("--- discordReady(", event, ") ---")
}

func discordChannelCreate(session *discord.Session, event *discord.ChannelCreate) {
	Log.Trace("--- discordChannelCreate(", event, ") ---")
}

func discordChannelUpdate(session *discord.Session, event *discord.ChannelUpdate) {
	Log.Trace("--- discordChannelUpdate(", event, ") ---")
}

func discordChannelDelete(session *discord.Session, event *discord.ChannelDelete) {
	Log.Trace("--- discordChannelDelete(", event, ") ---")
}

func discordGuildUpdate(session *discord.Session, event *discord.GuildUpdate) {
	Log.Trace("--- discordGuildUpdate(", event, ") ---")
}

func discordGuildBanAdd(session *discord.Session, event *discord.GuildBanAdd) {
	Log.Trace("--- discordGuildBanAdd(", event, ") ---")
}

func discordGuildBanRemove(session *discord.Session, event *discord.GuildBanRemove) {
	Log.Trace("--- discordGuildBanRemove(", event, ") ---")
}

func discordGuildMemberAdd(session *discord.Session, event *discord.GuildMemberAdd) {
	Log.Trace("--- discordGuildMemberAdd(", event, ") ---")
}

func discordGuildMemberRemove(session *discord.Session, event *discord.GuildMemberRemove) {
	Log.Trace("--- discordGuildMemberRemove(", event, ") ---")
}

func discordGuildRoleCreate(session *discord.Session, event *discord.GuildRoleCreate) {
	Log.Trace("--- discordGuildRoleCreate(", event, ") ---")
}

func discordGuildRoleUpdate(session *discord.Session, event *discord.GuildRoleUpdate) {
	Log.Trace("--- discordGuildRoleUpdate(", event, ") ---")
}

func discordGuildRoleDelete(session *discord.Session, event *discord.GuildRoleDelete) {
	Log.Trace("--- discordGuildRoleDelete(", event, ") ---")
}

func discordGuildEmojisUpdate(session *discord.Session, event *discord.GuildEmojisUpdate) {
	Log.Trace("--- discordGuildEmojisUpdate(", event, ") ---")
}

func discordUserUpdate(session *discord.Session, event *discord.UserUpdate) {
	Log.Trace("--- discordUserUpdate(", event, ") ---")
}

func discordVoiceStateUpdate(session *discord.Session, event *discord.VoiceStateUpdate) {
	Log.Trace("--- discordVoiceStateUpdate(", event, ") ---")
}

func discordMessageCreate(session *discord.Session, event *discord.MessageCreate) {
	Log.Trace("--- discordMessageCreate(", event, ") ---")
}

func discordMessageDelete(session *discord.Session, event *discord.MessageDelete) {
	Log.Trace("--- discordMessageDelete(", event, ") ---")
}

func discordMessageDeleteBulk(session *discord.Session, event *discord.MessageDeleteBulk) {
	Log.Trace("--- discordMessageDeleteBulkage(", event, ") ---")
}

func discordMessageUpdate(session *discord.Session, event *discord.MessageUpdate) {
	Log.Trace("--- discordMessageUpdate(", event, ") ---")
}

func discordMessageReactionAdd(session *discord.Session, event *discord.MessageReactionAdd) {
	Log.Trace("--- discordMessageReactionAdd(", event, ") ---")
}

func discordMessageReactionRemove(session *discord.Session, event *discord.MessageReactionRemove) {
	Log.Trace("--- discordMessageReactionRemove(", event, ") ---")
}

func discordMessageReactionRemoveAll(session *discord.Session, event *discord.MessageReactionRemoveAll) {
	Log.Trace("--- discordMessageReactionRemoveAll(", event, ") ---")
}