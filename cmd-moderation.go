package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//Moderator commands
func commandPurge(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	amount, err := strconv.Atoi(args[0])
	if err != nil {
		return NewErrorEmbed("Purge Error", "``"+args[0]+"`` is not a valid number.")
	}
	if amount <= 0 || amount > 100 {
		return NewErrorEmbed("Purge Error", "Amount of messages to purge must be between 1 and 100.")
	}

	messages, err := botData.DiscordSession.ChannelMessages(env.Channel.ID, amount, env.Message.ID, "", "")
	if err != nil {
		return NewErrorEmbed("Purge Error", "An error occurred fetching the last "+args[0]+" messages.")
	}

	messageIDs := make([]string, 0)

	if len(args) > 1 {
		for i := 0; i < len(messages); i++ {
			for j := 0; j < len(env.Message.Mentions); j++ {
				if env.Message.Mentions[j].ID == messages[i].Author.ID {
					messageIDs = append(messageIDs, messages[i].ID)
					break
				}
			}
		}

		err = botData.DiscordSession.ChannelMessagesBulkDelete(env.Channel.ID, messageIDs)
		if err != nil {
			return NewErrorEmbed("Purge Error", "An error occurred deleting the last "+args[0]+" messages from the specified user(s).")
		}

		return NewGenericEmbed("Purge", "Successfully purged the last "+args[0]+" messages from the specified user(s).")
	}

	for i := 0; i < len(messages); i++ {
		messageIDs = append(messageIDs, messages[i].ID)
	}

	err = botData.DiscordSession.ChannelMessagesBulkDelete(env.Channel.ID, messageIDs)
	if err != nil {
		return NewErrorEmbed("Purge Error", "An error occurred deleting the last "+args[0]+" messages.")
	}

	return NewGenericEmbed("Purge", "Successfully purged the last "+args[0]+" messages.")
}
func commandKick(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(env.Message.Mentions) == 0 {
		NewErrorEmbed("Kick Error", "You must specify which user(s) to kick from the server.")
	}

	reasonMessage := ""
	usersToKick := make([]string, 0)
	for i, part := range args {
		if strings.HasPrefix(part, "<@") && strings.HasSuffix(part, ">") {
			usersToKick = append(usersToKick, strings.TrimRight(strings.TrimLeft(part, "<@!"), ">"))
			continue
		}
		reasonMessage = strings.Join(args[i:], " ")
		break
	}
	if len(usersToKick) == 0 {
		return NewErrorEmbed("Kick Error", "You must specify which user(s) to kick.")
	}

	if reasonMessage == "" {
		for i := range usersToKick {
			err := botData.DiscordSession.GuildMemberDelete(env.Guild.ID, usersToKick[i])
			if err != nil {
				return NewErrorEmbed("Kick Error", "An error occurred kicking <@"+usersToKick[i]+">. Please consider manually kicking and report this issue to a developer.")
			}
		}
		return NewGenericEmbed("Kick", "Successfully kicked the selected user(s).")
	}
	for i := range usersToKick {
		err := botData.DiscordSession.GuildMemberDeleteWithReason(env.Guild.ID, usersToKick[i], reasonMessage)
		if err != nil {
			return NewErrorEmbed("Kick Error", "An error occurred kicking <@"+usersToKick[i]+">. Please consider manually kicking and report this issue to a developer.")
		}
	}
	return NewGenericEmbed("Kick", "Successfully kicked the selected user(s) for the following reason:\n**"+reasonMessage+"**")
}
func commandBan(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(env.Message.Mentions) == 0 {
		return NewErrorEmbed("Ban Error", "You must specify which user(s) to ban from the server.")
	}

	reasonMessage := ""
	usersToBan := make([]string, 0)
	messagesDaysToDelete := 0
	for i, part := range args {
		if i == 0 {
			days, err := strconv.Atoi(part)
			if err == nil {
				messagesDaysToDelete = days
				continue
			} else {
				return NewErrorEmbed("Ban Error", "``"+part+"`` is not a valid number.")
			}
		}
		if strings.HasPrefix(part, "<@") && strings.HasSuffix(part, ">") {
			usersToBan = append(usersToBan, strings.TrimRight(strings.TrimLeft(part, "<@!"), ">"))
			continue
		}
		reasonMessage = strings.Join(args[i:], " ")
		break
	}
	if len(usersToBan) == 0 {
		return NewErrorEmbed("Ban Error", "You must specify which user(s) to ban.")
	}
	if messagesDaysToDelete > 7 {
		return NewErrorEmbed("Ban Error", "You may only delete up to and including 7 days worth of messages from these users.")
	}

	if reasonMessage == "" {
		for i := range usersToBan {
			err := botData.DiscordSession.GuildBanCreate(env.Guild.ID, usersToBan[i], messagesDaysToDelete)
			if err != nil {
				return NewErrorEmbed("Ban Error", "An error occurred banning <@"+usersToBan[i]+">. Please consider manually banning and report this issue to a developer.")
			}
		}
		return NewGenericEmbed("Ban", "Successfully banned the selected user(s).")
	}
	for i := range usersToBan {
		err := botData.DiscordSession.GuildBanCreateWithReason(env.Guild.ID, usersToBan[i], reasonMessage, messagesDaysToDelete)
		if err != nil {
			return NewErrorEmbed("Ban Error", "An error occurred banning <@"+usersToBan[i]+">. Please consider manually banning and report this issue to a developer.")
		}
	}
	return NewGenericEmbed("Ban", "Successfully banned the selected user(s) for the following reason:\n**"+reasonMessage+"**")
}

func commandSettingsServer(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "joinmsg":
		guildSettings[env.Guild.ID].UserJoinMessage = strings.Join(args[1:], " ")
		guildSettings[env.Guild.ID].UserJoinMessageChannel = env.Channel.ID
		return NewGenericEmbed("Server Settings - Join Message", "Successfully set the join message to this channel.")
	case "leavemsg":
		guildSettings[env.Guild.ID].UserLeaveMessage = strings.Join(args[1:], " ")
		guildSettings[env.Guild.ID].UserLeaveMessageChannel = env.Channel.ID
		return NewGenericEmbed("Server Settings - Leave Message", "Successfully set the leave message to this channel.")
	case "reset":
		switch args[1] {
		case "joinmsg":
			guildSettings[env.Guild.ID].UserJoinMessage = ""
			guildSettings[env.Guild.ID].UserJoinMessageChannel = ""
		case "leavemsg":
			guildSettings[env.Guild.ID].UserLeaveMessage = ""
			guildSettings[env.Guild.ID].UserLeaveMessageChannel = ""
		}
	}
	return NewErrorEmbed("Server Settings Error", "Error finding the setting "+args[0]+".")
}