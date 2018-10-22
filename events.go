package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	defer recoverPanic()

	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: "+fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}

	go handleMessage(session, message, false)
}
func discordMessageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	defer recoverPanic()

	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: "+fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}

	go handleMessage(session, message, true)
}
func discordMessageDelete(session *discordgo.Session, event *discordgo.MessageDelete) {
	defer recoverPanic()

	message := event //Make it easier to keep track of what's happening

	guildChannel, err := session.Channel(message.ChannelID)
	if err == nil {
		guildID := guildChannel.GuildID

		guild, err := session.Guild(guildID)
		if err == nil {
			debugLog("[Deleted]["+guild.Name+" - #"+guildChannel.Name+"]: (Guild: "+guildID+", Channel: "+message.ChannelID+", Message: "+message.ID+")", false)

			_, guildFound := guildData[guildID]
			if guildFound {
				guildData[guildID].Lock()
				defer guildData[guildID].Unlock()

				_, messageFound := guildData[guildID].Queries[message.ID]
				if messageFound {
					debugLog("> Deleting message...", false)
					session.ChannelMessageDelete(message.ChannelID, guildData[guildID].Queries[message.ID].ResponseMessageID) //Delete the query response message
					guildData[guildID].Queries[message.ID] = nil                                                              //Remove the message from the query list
				} else {
					debugLog("> Error finding deleted message in queries list", false)
				}
			} else {
				debugLog("> Error finding guild for deleted message", false)
			}
		}
	} else {
		debugLog("> Error finding channel for deleted message", false)
	}
}
func discordMessageDeleteBulk(session *discordgo.Session, event *discordgo.MessageDeleteBulk) {
	defer recoverPanic()

	messages := event.Messages
	channelID := event.ChannelID

	guildChannel, err := session.Channel(channelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			guildData[guildID].Lock()
			defer guildData[guildID].Unlock()

			for i := 0; i > len(messages); i++ {
				debugLog("[D] ID: "+messages[i], false)
				_, messageFound := guildData[guildID].Queries[messages[i]]
				if messageFound {
					debugLog("> Deleting message...", false)
					session.ChannelMessageDelete(channelID, guildData[guildID].Queries[messages[i]].ResponseMessageID) //Delete the query response message
					guildData[guildID].Queries[messages[i]] = nil                                                      //Remove the message from the query list
				} else {
					debugLog("> Error finding deleted message in queries list", false)
				}
			}
		}
	}
}

func discordChannelCreate(session *discordgo.Session, channel *discordgo.ChannelCreate) {
	settings, guildFound := guildSettings[channel.GuildID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.ChannelCreate {
			switch channel.Type {
			case discordgo.ChannelTypeGuildText:
				channelCreateEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Create").
					SetDescription("A text channel was created.").
					AddField("Channel Name", channel.Name).
					AddField("Position", strconv.Itoa(channel.Position+1)).
					SetColor(0x1C1C1C)

				if channel.ParentID != "" {
					parent, err := session.Channel(channel.ParentID)
					if err == nil && parent.Type == discordgo.ChannelTypeGuildCategory {
						channelCreateEmbed.AddField("Parent Category", parent.Name)
					}
				}

				channelCreateEmbed.InlineAllFields()

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelCreateEmbed.MessageEmbed)
			case discordgo.ChannelTypeGuildVoice:
				channelCreateEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Create").
					SetDescription("A voice channel was created.").
					AddField("Voice Channel Name", ":speaker: "+channel.Name).
					AddField("Bitrate", strconv.Itoa(channel.Bitrate)).
					AddField("Position", strconv.Itoa(channel.Position+1)).
					SetColor(0x1C1C1C)

				if channel.ParentID != "" {
					parent, err := session.Channel(channel.ParentID)
					if err == nil && parent.Type == discordgo.ChannelTypeGuildCategory {
						channelCreateEmbed.AddField("Parent Category", parent.Name)
					}
				}

				channelCreateEmbed.InlineAllFields()

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelCreateEmbed.MessageEmbed)
			case discordgo.ChannelTypeGuildCategory:
				channelCreateEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Create").
					SetDescription("A category was created.").
					AddField("Category Name", channel.Name).
					AddField("Position", strconv.Itoa(channel.Position+1)).
					InlineAllFields().
					SetColor(0x1C1C1C).MessageEmbed

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelCreateEmbed)
			}
		}
	}
}
func discordChannelUpdate(session *discordgo.Session, channel *discordgo.ChannelUpdate) {
	settings, guildFound := guildSettings[channel.GuildID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.ChannelUpdate {
			switch channel.Type {
			case discordgo.ChannelTypeGuildText:
				channelUpdateEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Update").
					SetDescription("A text channel was updated.").
					AddField("Channel Name", channel.Name).
					AddField("NSFW", strconv.FormatBool(channel.NSFW)).
					AddField("Position", strconv.Itoa(channel.Position+1)).
					SetColor(0x1C1C1C)

				if channel.Topic != "" {
					channelUpdateEmbed.AddField("Topic", channel.Topic)
				}
				if channel.ParentID != "" {
					parent, err := session.Channel(channel.ParentID)
					if err == nil && parent.Type == discordgo.ChannelTypeGuildCategory {
						channelUpdateEmbed.AddField("Parent Category", parent.Name)
					}
				}

				channelUpdateEmbed.InlineAllFields()

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelUpdateEmbed.MessageEmbed)
			case discordgo.ChannelTypeGuildVoice:
				channelUpdateEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Update").
					SetDescription("A voice channel was updated.").
					AddField("Voice Channel Name", ":speaker: "+channel.Name).
					AddField("Bitrate", strconv.Itoa(channel.Bitrate)).
					AddField("Position", strconv.Itoa(channel.Position+1)).
					SetColor(0x1C1C1C)

				if channel.ParentID != "" {
					parent, err := session.Channel(channel.ParentID)
					if err == nil && parent.Type == discordgo.ChannelTypeGuildCategory {
						channelUpdateEmbed.AddField("Parent Category", parent.Name)
					}
				}

				channelUpdateEmbed.InlineAllFields()

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelUpdateEmbed.MessageEmbed)
			case discordgo.ChannelTypeGuildCategory:
				channelUpdateEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Update").
					SetDescription("A category was updated.").
					AddField("Category Name", channel.Name).
					AddField("Position", strconv.Itoa(channel.Position+1)).
					InlineAllFields().
					SetColor(0x1C1C1C).MessageEmbed

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelUpdateEmbed)
			}
		}
	}
}
func discordChannelDelete(session *discordgo.Session, channel *discordgo.ChannelDelete) {
	settings, guildFound := guildSettings[channel.GuildID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.ChannelDelete {
			switch channel.Type {
			case discordgo.ChannelTypeGuildText:
				channelDeleteEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Delete").
					SetDescription("A text channel was deleted.").
					AddField("Channel Name", channel.Name).
					SetColor(0x1C1C1C)

				if channel.ParentID != "" {
					parent, err := session.Channel(channel.ParentID)
					if err == nil && parent.Type == discordgo.ChannelTypeGuildCategory {
						channelDeleteEmbed.AddField("Parent Category", parent.Name)
					}
				}

				channelDeleteEmbed.InlineAllFields()

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelDeleteEmbed.MessageEmbed)
			case discordgo.ChannelTypeGuildVoice:
				channelDeleteEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Delete").
					SetDescription("A voice channel was deleted.").
					AddField("Voice Channel Name", ":speaker: "+channel.Name).
					SetColor(0x1C1C1C)

				if channel.ParentID != "" {
					parent, err := session.Channel(channel.ParentID)
					if err == nil && parent.Type == discordgo.ChannelTypeGuildCategory {
						channelDeleteEmbed.AddField("Parent Category", parent.Name)
					}
				}

				channelDeleteEmbed.InlineAllFields()

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelDeleteEmbed.MessageEmbed)
			case discordgo.ChannelTypeGuildCategory:
				channelDeleteEmbed := NewEmbed().
					SetTitle("Logging Event - Channel Delete").
					SetDescription("A category was deleted.").
					AddField("Category Name", channel.Name).
					InlineAllFields().
					SetColor(0x1C1C1C).MessageEmbed

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, channelDeleteEmbed)
			}
		}
	}
}
func discordGuildUpdate(session *discordgo.Session, guild *discordgo.GuildUpdate) {
	settings, guildFound := guildSettings[guild.ID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.GuildUpdate {
			verificationLevel := "None"
			switch guild.VerificationLevel {
			case discordgo.VerificationLevelLow:
				verificationLevel = "Low"
			case discordgo.VerificationLevelMedium:
				verificationLevel = "Medium"
			case discordgo.VerificationLevelHigh:
				verificationLevel = "High"
			}
			afkChannel := "None"
			if guild.AfkChannelID != "" {
				channel, err := botData.DiscordSession.Channel(guild.AfkChannelID)
				if err == nil && channel.Type == discordgo.ChannelTypeGuildVoice {
					afkChannel = ":speaker: " + channel.Name
				}
			}

			session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, NewEmbed().
				SetTitle("Logging Event - Guild Update").
				SetDescription("The guild was updated.").
				AddField("Guild Name", guild.Name).
				AddField("Guild Region", guild.Region).
				AddField("AFK Channel", afkChannel).
				AddField("Owner", "<@"+guild.OwnerID+">").
				AddField("Verification Level", verificationLevel).
				AddField("Embeds Enabled", strconv.FormatBool(guild.EmbedEnabled)).
				AddField("Large Member Count", strconv.FormatBool(guild.Large)).
				SetImage("https://cdn.discordapp.com/icons/"+guild.ID+"/"+guild.Icon+".jpg").
				InlineAllFields().
				SetColor(0x1C1C1C).MessageEmbed,
			)
		}
	}
}
func discordGuildBanAdd(session *discordgo.Session, guild *discordgo.GuildBanAdd) {
	settings, guildFound := guildSettings[guild.GuildID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.GuildBanAdd {
			session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, NewEmbed().
				SetTitle("Logging Event - Ban Add").
				SetDescription("A member was banned from the server.").
				AddField("User ID", guild.User.ID).
				AddField("Username", guild.User.Username+"#"+guild.User.Discriminator).
				SetImage(guild.User.AvatarURL("2048")).
				InlineAllFields().
				SetColor(0x1C1C1C).MessageEmbed,
			)
		}
	}
}
func discordGuildBanRemove(session *discordgo.Session, guild *discordgo.GuildBanRemove) {
	settings, guildFound := guildSettings[guild.GuildID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.GuildBanRemove {
			session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, NewEmbed().
				SetTitle("Logging Event - Ban Remove").
				SetDescription("A member was unbanned from the server.").
				AddField("User ID", guild.User.ID).
				AddField("Username", guild.User.Username+"#"+guild.User.Discriminator).
				SetImage(guild.User.AvatarURL("2048")).
				InlineAllFields().
				SetColor(0x1C1C1C).MessageEmbed,
			)
		}
	}
}
func discordGuildMemberAdd(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	_, guildFound := guildSettings[member.GuildID]
	if guildFound {
		if guildSettings[member.GuildID].UserJoinMessage != "" && guildSettings[member.GuildID].UserJoinMessageChannel != "" {
			message := guildSettings[member.GuildID].UserJoinMessage
			message = strings.Replace(message, "{user}", member.User.Username, -1)
			message = strings.Replace(message, "{user-mention}", "<@"+member.User.ID+">", -1)
			message = strings.Replace(message, "{user-id}", member.User.ID, -1)
			message = strings.Replace(message, "{user-discriminator}", member.User.Discriminator, -1)

			session.ChannelMessageSend(guildSettings[member.GuildID].UserJoinMessageChannel, message)
		}

		if guildSettings[member.GuildID].LogSettings.LoggingEnabled && guildSettings[member.GuildID].LogSettings.LoggingEvents.GuildMemberAdd {
			joinedAt := member.JoinedAt
			joinedAtTimeFormatted := ""
			joinedAtTime, err := joinedAt.Parse()
			if err != nil {
				joinedAtTimeFormatted = string(joinedAt)
			} else {
				joinedAtMonth := joinedAtTime.Month().String()
				joinedAtDay := joinedAtTime.Day()
				joinedAtYear := joinedAtTime.Year()
				joinedAtHour := joinedAtTime.Hour()
				joinedAtMinute := joinedAtTime.Minute()
				joinedAtSecond := joinedAtTime.Second()
				joinedAtTimeFormatted = joinedAtMonth + " " + strconv.Itoa(joinedAtDay) + ", " + strconv.Itoa(joinedAtYear) + " at " + strconv.Itoa(joinedAtHour) + ":" + strconv.Itoa(joinedAtMinute) + ":" + strconv.Itoa(joinedAtSecond)
			}

			session.ChannelMessageSendEmbed(guildSettings[member.GuildID].LogSettings.LoggingChannel, NewEmbed().
				SetTitle("Logging Event - User Joined").
				SetDescription("A new member joined the server.").
				AddField("Joined At", joinedAtTimeFormatted).
				AddField("User ID", member.User.ID).
				AddField("Username", member.User.Username+"#"+member.User.Discriminator).
				AddField("Verified Account", strconv.FormatBool(member.User.Verified)).
				AddField("Multi-Factor Authentication", strconv.FormatBool(member.User.MFAEnabled)).
				AddField("Bot", strconv.FormatBool(member.User.Bot)).
				SetImage(member.User.AvatarURL("2048")).
				InlineAllFields().
				SetColor(0x1C1C1C).MessageEmbed,
			)
		}
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

		if guildSettings[member.GuildID].LogSettings.LoggingEnabled && guildSettings[member.GuildID].LogSettings.LoggingEvents.GuildMemberRemove {
			joinedAt := member.JoinedAt
			joinedAtTimeFormatted := ""
			joinedAtTime, err := joinedAt.Parse()
			if err != nil {
				joinedAtTimeFormatted = string(joinedAt)
			} else {
				joinedAtMonth := joinedAtTime.Month().String()
				joinedAtDay := joinedAtTime.Day()
				joinedAtYear := joinedAtTime.Year()
				joinedAtHour := joinedAtTime.Hour()
				joinedAtMinute := joinedAtTime.Minute()
				joinedAtSecond := joinedAtTime.Second()
				joinedAtTimeFormatted = joinedAtMonth + " " + strconv.Itoa(joinedAtDay) + ", " + strconv.Itoa(joinedAtYear) + " at " + strconv.Itoa(joinedAtHour) + ":" + strconv.Itoa(joinedAtMinute) + ":" + strconv.Itoa(joinedAtSecond)
			}

			session.ChannelMessageSendEmbed(guildSettings[member.GuildID].LogSettings.LoggingChannel, NewEmbed().
				SetTitle("Logging Event - User Left").
				SetDescription("A member left the server.").
				AddField("Joined At", joinedAtTimeFormatted).
				AddField("User ID", member.User.ID).
				AddField("Username", member.User.Username+"#"+member.User.Discriminator).
				AddField("Verified Account", strconv.FormatBool(member.User.Verified)).
				AddField("Multi-Factor Authentication", strconv.FormatBool(member.User.MFAEnabled)).
				AddField("Bot", strconv.FormatBool(member.User.Bot)).
				SetImage(member.User.AvatarURL("2048")).
				InlineAllFields().
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
	settings, guildFound := guildSettings[voiceState.GuildID]
	if guildFound {
		if settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.VoiceStateUpdate {
			if voiceState.ChannelID == "" {
				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, NewEmbed().
					SetTitle("Logging Event - Voice State Update").
					SetDescription("A voice state was updated.").
					AddField("User", "<@"+voiceState.UserID+">").
					AddField("Voice Channel", "None").
					InlineAllFields().
					SetColor(0x1C1C1C).MessageEmbed,
				)
			} else {
				voiceChannel, err := session.Channel(voiceState.ChannelID)
				if err != nil {
					return
				}

				session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, NewEmbed().
					SetTitle("Logging Event - Voice State Update").
					SetDescription("A voice state was updated.").
					AddField("User", "<@"+voiceState.UserID+">").
					AddField("Voice Channel", ":speaker: "+voiceChannel.Name).
					AddField("State",
						"Suppress: **"+strconv.FormatBool(voiceState.Suppress)+"**\n"+
							"Self Mute: **"+strconv.FormatBool(voiceState.SelfMute)+"**\n"+
							"Self Deafen: **"+strconv.FormatBool(voiceState.SelfDeaf)+"**\n"+
							"Server Mute: **"+strconv.FormatBool(voiceState.Mute)+"**\n"+
							"Server Deafen: **"+strconv.FormatBool(voiceState.Deaf)+"**",
					).
					InlineAllFields().
					SetColor(0x1C1C1C).MessageEmbed,
				)

			}
		}
	}
}
