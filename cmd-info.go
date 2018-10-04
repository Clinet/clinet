package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func commandBotInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	guildCount := len(botData.DiscordSession.State.Guilds)
	commandCount := 0
	for _, command := range botData.Commands {
		if command.IsAlternateOf == "" {
			commandCount++
		}
	}

	return NewEmbed().
		SetAuthor(botData.BotName, botData.DiscordSession.State.User.AvatarURL("2048")).
		AddField("Bot Owner", "<@!"+botData.BotOwnerID+">").
		AddField("Guild Count", strconv.Itoa(guildCount)).
		AddField("Default Prefix", botData.CommandPrefix).
		AddField("Command Count", strconv.Itoa(commandCount)).
		AddField("Disabled Wolfram|Alpha Pods", strings.Join(botData.BotOptions.WolframDeniedPods, ", ")).
		AddField("Debug Mode", strconv.FormatBool(botData.DebugMode)).
		InlineAllFields().
		SetColor(0x1C1C1C).MessageEmbed
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
	if len(env.Message.Mentions) > 0 {
		user = env.Message.Mentions[0]
	}

	creationDate := ""
	creationTime, err := CreationTime(user.ID)
	if err != nil {
		creationDate = "Unable to find creation date"
	} else {
		creationDate = creationTime.Format("01/02/2006 15:04:05 MST")
	}

	userInfoEmbed := NewEmbed().
		SetAuthor(user.Username+"#"+user.Discriminator, user.AvatarURL("2048")).
		AddField("User ID", user.ID).
		AddField("Discriminator", user.Discriminator).
		AddField("Creation Date", creationDate).
		AddField("Bot", strconv.FormatBool(user.Bot)).
		SetColor(0x1C1C1C)

	if userSettings, found := userSettings[user.ID]; found {
		if userSettings.AboutMe != "" {
			userInfoEmbed.AddField("About Me", userSettings.AboutMe)
		}
	}

	userInfoEmbed.InlineAllFields()

	return userInfoEmbed.MessageEmbed
}
