package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//RoleMe stores a roleme event
type RoleMe struct {
	Triggers      []string //An array of messages to trigger this roleme event
	AddRoles      []string //An array of roles to add
	RemoveRoles   []string //An array of roles to remove
	CaseSensitive bool     //Whether or not the trigger message should be case-sensitive
	ChannelIDs    []string //An array of channel IDs to apply this roleme event to
}

func commandRoleMe(args []CommandArgument, env *CommandEnvironment) *discordgo.MessageEmbed {
	rolesToAdd := make([]string, 0)
	rolesToRemove := make([]string, 0)
	channelIDs := make([]string, 0)
	triggers := make([]string, 0)
	caseSensitive := false

	for _, arg := range args {
		switch strings.ToLower(arg.Name) {
		case "addrole", "roleadd":
			if arg.Value == "" {
				return NewErrorEmbed("RoleMe Error", "You must supply a value to the addrole argument.")
			}
			role, err := getRole(env.Guild.ID, arg.Value)
			if err != nil {
				return NewErrorEmbed("RoleMe Error", "Error finding role %s.", arg.Value)
			}
			if isStrInSlice(rolesToAdd, role.ID) {
				return NewErrorEmbed("RoleMe Error", "You cannot specify the same role to add twice.")
			}
			if isStrInSlice(rolesToRemove, role.ID) {
				return NewErrorEmbed("RoleMe Error", "You cannot specify a role to add if the role is already specified to be removed.")
			}
			rolesToAdd = append(rolesToAdd, role.ID)
		case "removerole", "roleremove", "deleterole", "roledelete":
			if arg.Value == "" {
				return NewErrorEmbed("RoleMe Error", "You must supply a value to the removerole argument.")
			}
			role, err := getRole(env.Guild.ID, arg.Value)
			if err != nil {
				return NewErrorEmbed("RoleMe Error", "Error finding role %s.", arg.Value)
			}
			if isStrInSlice(rolesToRemove, role.ID) {
				return NewErrorEmbed("RoleMe Error", "You cannot specify the same role to remove twice.")
			}
			if isStrInSlice(rolesToAdd, role.ID) {
				return NewErrorEmbed("RoleMe Error", "You cannot specify a role to remove if the role is already specified to be added.")
			}
			rolesToRemove = append(rolesToRemove, role.ID)
		case "casesensitive":
			switch arg.Value {
			case "true", "t", "1", "yes", "y", "":
				caseSensitive = true
			case "false", "f", "0", "no", "n":
				caseSensitive = false
			}
		case "channel":
			if arg.Value == "" {
				return NewErrorEmbed("RoleMe Error", "You must supply a value to the channel argument.")
			}
			channel, err := getChannel(env.Guild.ID, arg.Value)
			if err != nil {
				return NewErrorEmbed("RoleMe Error", "Error finding channel %s.", arg.Value)
			}
			if isStrInSlice(channelIDs, channel.ID) {
				return NewErrorEmbed("RoleMe Error", "You cannot specify the same channel twice.")
			}
			channelIDs = append(channelIDs, channel.ID)
		case "trigger", "message", "msg":
			if arg.Value == "" {
				return NewErrorEmbed("RoleMe Error", "You must supply a value to the trigger argument.")
			}
			if isStrInSlice(triggers, arg.Value) {
				return NewErrorEmbed("RoleMe Error", "You cannot specify the same trigger twice.")
			}
			triggers = append(triggers, arg.Value)
		default:
			return NewErrorEmbed("RoleMe Error", "Unknown argument ``%s``.", arg.Name)
		}
	}

	if len(rolesToAdd) == 0 && len(rolesToRemove) == 0 {
		return NewErrorEmbed("RoleMe Error", "You must specify either one or more roles to add or one or more roles to remove.")
	}
	if len(triggers) == 0 {
		return NewErrorEmbed("RoleMe Error", "You must specify one or more triggers to trigger this roleme event.")
	}

	newRoleMe := &RoleMe{
		Triggers:      triggers,
		AddRoles:      rolesToAdd,
		RemoveRoles:   rolesToRemove,
		CaseSensitive: caseSensitive,
		ChannelIDs:    channelIDs,
	}

	for _, roleMe := range guildSettings[env.Guild.ID].RoleMeList {
		for _, trigger := range roleMe.Triggers {
			for _, newTrigger := range newRoleMe.Triggers {
				if trigger == newTrigger {
					return NewErrorEmbed("RoleMe Error", "The trigger ``%s`` already exists!", trigger)
				}
			}
		}
	}

	guildSettings[env.Guild.ID].RoleMeList = append(guildSettings[env.Guild.ID].RoleMeList, newRoleMe)
	return NewGenericEmbed("RoleMe", "Added the roleme event successfully!")
}

func handleRoleMe(roleMe *RoleMe, guildID, channelID, userID string) {
	if len(roleMe.ChannelIDs) > 0 {
		channelFound := false
		for _, roleMeChannelID := range roleMe.ChannelIDs {
			if roleMeChannelID == channelID {
				channelFound = true
				break
			}
		}
		if channelFound == false {
			return
		}
	}

	errCount := 0
	successCount := 0
	for _, roleToAdd := range roleMe.AddRoles {
		err := botData.DiscordSession.GuildMemberRoleAdd(guildID, userID, roleToAdd)
		if err != nil {
			errCount++
		} else {
			successCount++
		}
	}
	for _, roleToRemove := range roleMe.RemoveRoles {
		err := botData.DiscordSession.GuildMemberRoleRemove(guildID, userID, roleToRemove)
		if err != nil {
			errCount++
		} else {
			successCount++
		}
	}

	if errCount == 0 {
		botData.DiscordSession.ChannelMessageSendEmbed(channelID, NewGenericEmbed("RoleMe", "Edited your roles successfully!"))
	} else if errCount < successCount {
		botData.DiscordSession.ChannelMessageSendEmbed(channelID, NewGenericEmbed("RoleMe", "There were some errors editing your roles, but there were more successes!"))
	} else {
		botData.DiscordSession.ChannelMessageSendEmbed(channelID, NewErrorEmbed("RoleMe Error", "There were some errors editing your roles. :c"))
	}
}

func getRole(guildID, role string) (*discordgo.Role, error) {
	guildRoles, err := botData.DiscordSession.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	retRole := &discordgo.Role{}

	if strings.HasPrefix(role, "<&") && strings.HasSuffix(role, ">") {
		roleID := role
		roleID = strings.TrimPrefix(roleID, "<&")
		roleID = strings.TrimSuffix(roleID, ">")

		roleFound := false
		for _, guildRole := range guildRoles {
			if guildRole.ID == roleID {
				retRole = guildRole
				roleFound = true
				break
			}
		}
		if roleFound == false {
			return nil, fmt.Errorf("error finding role by ID %s", roleID)
		}
	} else {
		roleFound := false
		for _, guildRole := range guildRoles {
			if guildRole.Name == role {
				retRole = guildRole
				roleFound = true
				break
			}
		}
		if roleFound == false {
			return nil, fmt.Errorf("error finding role by name %s", role)
		}
	}

	return retRole, nil
}

func getChannel(guildID, channel string) (*discordgo.Channel, error) {
	guildChannels, err := botData.DiscordSession.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	retChannel := &discordgo.Channel{}

	if strings.HasPrefix(channel, "<#") && strings.HasSuffix(channel, ">") {
		channelID := channel
		channelID = strings.TrimPrefix(channelID, "<#")
		channelID = strings.TrimSuffix(channelID, ">")

		channelFound := false
		for _, guildChannel := range guildChannels {
			if guildChannel.ID == channelID {
				retChannel = guildChannel
				channelFound = true
				break
			}
		}
		if channelFound == false {
			return nil, fmt.Errorf("error finding channel by ID %s", channelID)
		}
	} else {
		channelFound := false
		for _, guildChannel := range guildChannels {
			if guildChannel.Name == channel {
				retChannel = guildChannel
				channelFound = true
				break
			}
		}
		if channelFound == false {
			return nil, fmt.Errorf("error finding channel by name %s", channel)
		}
	}

	return retChannel, nil
}

func isStrInSlice(slice []string, str string) bool {
	for _, value := range slice {
		if value == str {
			return true
		}
	}
	return false
}
