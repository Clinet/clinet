package main

import (
	"strconv"
	"strings"
	"time"

	"4d63.com/tz"
	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
)

func commandBotInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	guildCount := len(botData.DiscordSession.State.Guilds)
	commandCount := 0
	for _, command := range botData.Commands {
		if command.IsAlternateOf == "" {
			commandCount++
		}
	}

	botEmbed := NewEmbed().
		SetAuthor(botData.BotName, botData.DiscordSession.State.User.AvatarURL("2048")).
		AddField("Bot Owner", "<@!"+botData.BotOwnerID+">").
		AddField("Guild Count", strconv.Itoa(guildCount)).
		AddField("Default Prefix", botData.CommandPrefix).
		AddField("Command Count", strconv.Itoa(commandCount)).
		AddField("Uptime", humanize.Time(uptime)).
		AddField("Debug Mode", strconv.FormatBool(botData.DebugMode)).
		InlineAllFields().
		SetColor(0x1C1C1C)

	enabledFeatures := make([]string, 0)
	if botData.BotOptions.UseCustomResponses {
		enabledFeatures = append(enabledFeatures, "Custom Responses")
	}
	if botData.BotOptions.UseDuckDuckGo {
		enabledFeatures = append(enabledFeatures, "DuckDuckGo")
	}
	if botData.BotOptions.UseGitHub {
		enabledFeatures = append(enabledFeatures, "GitHub")
	}
	if botData.BotOptions.UseImgur {
		enabledFeatures = append(enabledFeatures, "Imgur")
	}
	if botData.BotOptions.UseSoundCloud {
		enabledFeatures = append(enabledFeatures, "SoundCloud")
	}
	if botData.BotOptions.UseSpotify {
		enabledFeatures = append(enabledFeatures, "Spotify")
	}
	if botData.BotOptions.UseWolframAlpha {
		enabledFeatures = append(enabledFeatures, "Wolfram|Alpha")
	}
	if botData.BotOptions.UseXKCD {
		enabledFeatures = append(enabledFeatures, "xkcd")
	}
	if botData.BotOptions.UseYouTube {
		enabledFeatures = append(enabledFeatures, "YouTube")
	}
	if len(enabledFeatures) > 0 {
		botEmbed.AddField("Enabled Features", strings.Join(enabledFeatures, ", "))
	}

	botEmbed.AddField("Reason for Downtime", DowntimeReason)

	return botEmbed.MessageEmbed
}

func commandServerInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	verificationLevel := "None"
	switch env.Guild.VerificationLevel {
	case discordgo.VerificationLevelLow:
		verificationLevel = "Low"
	case discordgo.VerificationLevelMedium:
		verificationLevel = "Medium"
	case discordgo.VerificationLevelHigh:
		verificationLevel = "High"
	}

	afkChannel := "None"
	if env.Guild.AfkChannelID != "" {
		channel, err := botData.DiscordSession.Channel(env.Guild.AfkChannelID)
		if err == nil && channel.Type == discordgo.ChannelTypeGuildVoice {
			afkChannel = ":speaker: " + channel.Name
		}
	}

	creationDate := ""
	creationTime, err := CreationTime(env.Guild.ID)
	if err != nil {
		creationDate = "Unable to find creation date"
	} else {
		creationDate = creationTime.Format("01/02/2006 15:04:05 MST")
	}

	roleCount := len(env.Guild.Roles)
	emojiCount := len(env.Guild.Emojis)
	channelCount := len(env.Guild.Channels)
	voiceStateCount := len(env.Guild.VoiceStates)

	guildIcon := "https://cdn.discordapp.com/icons/" + env.Guild.ID + "/" + env.Guild.Icon + ".jpg"

	return NewEmbed().
		SetAuthor(env.Guild.Name, guildIcon).
		AddField("Server ID", env.Guild.ID).
		AddField("Server Region", env.Guild.Region).
		AddField("Server Owner", "<@!"+env.Guild.OwnerID+">").
		AddField("Creation Date", creationDate).
		AddField("Verification Level", verificationLevel).
		AddField("AFK Voice Channel", afkChannel).
		AddField("AFK Timeout", strconv.Itoa(env.Guild.AfkTimeout)+" seconds").
		AddField("Member Count", strconv.Itoa(env.Guild.MemberCount)).
		AddField("Role Count", strconv.Itoa(roleCount)).
		AddField("Custom Emoji Count", strconv.Itoa(emojiCount)).
		AddField("Channel Count", strconv.Itoa(channelCount)).
		AddField("Voice State Count", strconv.Itoa(voiceStateCount)).
		InlineAllFields().
		SetColor(0x1C1C1C).MessageEmbed
}

func commandUserInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	user := env.User
	member := env.Member
	memberFound := true
	if len(env.Message.Mentions) > 0 {
		user = env.Message.Mentions[0]

		memberMention, err := botData.DiscordSession.GuildMember(env.Guild.ID, user.ID)
		if err != nil {
			memberFound = false
		}
		member = memberMention
	} else if len(args) > 0 {
		mention := args[0]

		userMention, err := botData.DiscordSession.User(mention)
		if err != nil {
			return NewErrorEmbed("User Info Error", "Invalid user ``"+mention+"``.")
		}
		user = userMention

		memberMention, err := botData.DiscordSession.GuildMember(env.Guild.ID, user.ID)
		if err != nil {
			memberFound = false
		}
		member = memberMention
	}

	timezone := userSettings[env.User.ID].Timezone
	if timezone == "" {
		return NewErrorEmbed("User Info Error", "Please set a timezone first!\n\nEx: ``"+env.BotPrefix+"user timezone America/New_York``")
	}
	location, err := tz.LoadLocation(timezone)
	if err != nil {
		return NewErrorEmbed("User Info Error", "You have an invalid timezone set, please set a new one first!\n\nEx: ``"+env.BotPrefix+"user timezone America/New_York``")
	}

	creationDate := ""
	creationTime, err := CreationTime(user.ID)
	if err != nil {
		creationDate = "Unable to find creation date"
	} else {
		creationDate = creationTime.In(location).Format("01/02/2006 15:04:05 MST")
	}

	userInfoEmbed := NewEmbed().
		SetColor(0x1C1C1C).
		SetFooter(user.ID)

	author := user.Username + "#" + user.Discriminator

	if memberFound {
		if member.Nick != "" {
			author += " AKA " + member.Nick
		}
	}

	userInfoEmbed.SetAuthor(author, user.AvatarURL("2048")).
		AddField("Creation Date", creationDate)

	if memberFound {
		joinedAtTime, err := member.JoinedAt.Parse()
		if err == nil {
			userInfoEmbed.AddField("Guild Join Date", joinedAtTime.In(location).Format("01/02/2006 15:04:05 MST"))
		}
	}

	userInfoEmbed.AddField("Bot", strconv.FormatBool(user.Bot))

	if memberFound {
		if len(member.Roles) > 0 {
			roles := make([]string, 0)
			for _, roleID := range member.Roles {
				role, err := botData.DiscordSession.State.Role(env.Guild.ID, roleID)
				if err == nil {
					roles = append(roles, role.Name)
				}
			}
			userInfoEmbed.AddField("Roles", strings.Join(roles, ", "))
		}
	}

	if memberFound {
		presence, err := botData.DiscordSession.State.Presence(env.Guild.ID, user.ID)
		if err == nil {
			status := ""
			switch presence.Status {
			case discordgo.StatusOnline:
				status = "Online"
			case discordgo.StatusOffline:
				status = "Offline"
			case discordgo.StatusIdle:
				status = "Idle"
			case discordgo.StatusDoNotDisturb:
				status = "Do Not Disturb"
			case discordgo.StatusInvisible:
				status = "Invisible"
			}
			if presence.Game != nil {
				gameName := presence.Game.Name
				if presence.Game.URL != "" {
					gameName = "[" + presence.Game.Name + "](" + presence.Game.URL + ")"
				}
				switch presence.Game.Type {
				case discordgo.GameTypeGame:
					status += ", playing " + gameName
				case discordgo.GameTypeStreaming:
					status += ", streaming " + gameName
				}
				if presence.Game.TimeStamps.StartTimestamp != 0 {
					status += " as of " + humanize.Time(time.Unix(presence.Game.TimeStamps.StartTimestamp, 0).In(location))
				}
			}
			userInfoEmbed.AddField("Presence", status)
		}
	}

	if userSettings, found := userSettings[user.ID]; found {
		if userSettings.AboutMe != "" {
			userInfoEmbed.AddField("About Me", userSettings.AboutMe)
		}
		if userSettings.Timezone != "" {
			userInfoEmbed.AddField("Timezone", userSettings.Timezone)
		}
		if userSettings.Socials.SwitchFC != "" || userSettings.Socials.NNID != "" || userSettings.Socials.PSN != "" || userSettings.Socials.Xbox != "" {
			socials := userSettings.Socials
			socialsField := ""

			if socials.SwitchFC != "" {
				socialsField += "**Switch Friend Code**: " + socials.SwitchFC + "\n"
			}
			if socials.NNID != "" {
				socialsField += "**NNID**: " + socials.NNID + "\n"
			}
			if socials.PSN != "" {
				socialsField += "**PSN**: " + socials.PSN + "\n"
			}
			if socials.Xbox != "" {
				socialsField += "**Xbox Live Gamertag**: " + socials.Xbox + "\n"
			}

			userInfoEmbed.AddField("Socials", socialsField)
		}
	}

	userInfoEmbed.InlineAllFields()

	return userInfoEmbed.MessageEmbed
}
