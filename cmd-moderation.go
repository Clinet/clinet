package main

import (
	"fmt"
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
			if strings.TrimRight(strings.TrimLeft(strings.TrimLeft(part, "<@"), "!"), ">") == env.User.ID {
				return NewErrorEmbed("Kick Error", "You can't kick yourself!")
			}
			usersToKick = append(usersToKick, strings.TrimRight(strings.TrimLeft(strings.TrimLeft(part, "<@"), "!"), ">"))
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
			if strings.TrimRight(strings.TrimLeft(strings.TrimLeft(part, "<@"), "!"), ">") == env.User.ID {
				return NewErrorEmbed("Kick Error", "You can't kick yourself!")
			}
			usersToBan = append(usersToBan, strings.TrimRight(strings.TrimLeft(strings.TrimLeft(part, "<@"), "!"), ">"))
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
func commandHackBan(args []CommandArgument, env *CommandEnvironment) *discordgo.MessageEmbed {
	reasonMessage := ""
	usersToBan := make([]string, 0)
	messagesDaysToDelete := 0

	for i := 0; i < len(args); i++ {
		switch args[i].Name {
		case "days":
			if args[i].Value == "" {
				return NewErrorEmbed("HackBan Error", "You must specify how many days of messages to delete if you use the ``-days`` argument.")
			}
			days, err := strconv.Atoi(args[i].Value)
			if err != nil {
				return NewErrorEmbed("HackBan Error", "Invalid days ``"+args[i].Value+"``.")
			}
			messagesDaysToDelete = days
		case "id":
			if args[i].Value == "" {
				return NewErrorEmbed("HackBan Error", "You must specify the ID of the user to hackban if you use the ``-id`` argument.")
			}
			usersToBan = append(usersToBan, args[i].Value)
		case "reason":
			if args[i].Value == "" {
				return NewErrorEmbed("HackBan Error", "You must specify the reason for hackbanning if you use the ``-reason`` argument.")
			}
			reasonMessage = args[i].Value
		}
	}

	if len(usersToBan) == 0 {
		return NewErrorEmbed("HackBan Error", "You must specify which user IDs to hackban.")
	}

	if reasonMessage == "" {
		reasonMessage = "Banned by " + env.User.Username + "#" + env.User.Discriminator + " using Clinet"
	}

	failedBans := make([]string, 0)
	failedErrors := make([]error, 0)
	for i := range usersToBan {
		err := botData.DiscordSession.GuildBanCreateWithReason(env.Guild.ID, usersToBan[i], reasonMessage, messagesDaysToDelete)
		if err != nil {
			failedBans = append(failedBans, usersToBan[i])
			failedErrors = append(failedErrors, err)
		}
	}

	resp := "Successfully hackbanned the specified user!"
	if len(usersToBan) > 1 {
		resp = "Successfully hackbanned the specified users!"
	}

	if len(failedBans) > 0 {
		resp = "There was an error hackbanning the following users:\n"
		for i := 0; i < len(failedBans); i++ {
			resp += fmt.Sprintf("\n- <!%s>: %v", failedBans[i], failedErrors[i])
		}
		return NewErrorEmbed("HackBan Error", resp)
	}
	return NewGenericEmbed("HackBan", resp)
}
