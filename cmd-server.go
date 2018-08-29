package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/structs"
)

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
	case "log":
		if len(args) < 2 {
			logHelpCmd := &Command{
				HelpText: "Sets the logging capabilities for this server.",
				RequiredArguments: []string{
					"setting (value(s))",
				},
				Arguments: []CommandArgument{
					{Name: "set", Description: "Sets the logging channel to the current channel", ArgType: "this"},
					{Name: "enable", Description: "Enables logging for the server (to this channel if not set), enabling any optionally specified events", ArgType: "this/event(s)"},
					{Name: "disable", Description: "Disables logging for the server, disabling any optionally specified events", ArgType: "this/event(s)"},
					{Name: "unset", Description: "Unsets the current logging channel and disables logging", ArgType: "this"},
					{Name: "events", Description: "Returns a list of available events to enable/disable", ArgType: "this"},
				},
			}
			return getCustomCommandUsage(logHelpCmd, "server log", "Server Settings - Log Help")
		}

		LoggingEventsTmp := &guildSettings[env.Guild.ID].LogSettings.LoggingEvents /*
			LoggingEventsTmp.ChannelCreate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.ChannelCreate
			LoggingEventsTmp.ChannelUpdate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.ChannelUpdate
			LoggingEventsTmp.ChannelDelete = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.ChannelDelete
			LoggingEventsTmp.GuildUpdate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildUpdate
			LoggingEventsTmp.GuildBanAdd = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildBanAdd
			LoggingEventsTmp.GuildBanRemove = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildBanRemove
			LoggingEventsTmp.GuildMemberAdd = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildMemberAdd
			LoggingEventsTmp.GuildMemberRemove = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildMemberRemove
			LoggingEventsTmp.GuildRoleCreate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildRoleCreate
			LoggingEventsTmp.GuildRoleUpdate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildRoleUpdate
			LoggingEventsTmp.GuildRoleDelete = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildRoleDelete
			LoggingEventsTmp.GuildEmojisUpdate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.GuildEmojisUpdate
			LoggingEventsTmp.UserUpdate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.UserUpdate
			LoggingEventsTmp.VoiceStateUpdate = guildSettings[env.Guild.ID].LogSettings.LoggingEvents.VoiceStateUpdate */

		switch args[1] {
		case "set":
			guildSettings[env.Guild.ID].LogSettings.LoggingChannel = env.Channel.ID
			return NewGenericEmbed("Server Settings - Log", "Successfully set the logging channel to this channel.")
		case "enable":
			guildSettings[env.Guild.ID].LogSettings.LoggingEnabled = true

			if len(args) == 3 {
				switch args[2] {
				case "all":
					events := structs.New(LoggingEventsTmp)
					fields := events.Fields()

					for _, event := range fields {
						err := event.Set(true)
						if err != nil {
							return NewErrorEmbed("Server Settings - Log", "Unable to enable all logging events.")
						}
					}

					guildSettings[env.Guild.ID].LogSettings.LoggingEvents = *LoggingEventsTmp

					if guildSettings[env.Guild.ID].LogSettings.LoggingChannel == "" {
						guildSettings[env.Guild.ID].LogSettings.LoggingChannel = env.Channel.ID
						return NewGenericEmbed("Server Settings - Log", "Successfully enabled all logging events and set the logging channel to this channel.")
					}

					return NewGenericEmbed("Server Settings - Log", "Successfully enabled all logging events.")
				case "recommended":
					guildSettings[env.Guild.ID].LogSettings.LoggingEvents = LogEventsRecommended

					if guildSettings[env.Guild.ID].LogSettings.LoggingChannel == "" {
						guildSettings[env.Guild.ID].LogSettings.LoggingChannel = env.Channel.ID
						return NewGenericEmbed("Server Settings - Log", "Successfully toggled all logging events to their recommended states and set the logging channel to this channel.")
					}

					return NewGenericEmbed("Server Settings - Log", "Successfully toggled all logging events to their recommended states.")
				}
			}

			eventsToEnable := make([]string, 0)
			if len(args) > 2 {
				eventsToEnable = args[2:]
			}
			eventsEnabled := make([]string, 0)
			eventsFailed := make([]string, 0)

			if len(eventsToEnable) > 0 {
				events := structs.New(LoggingEventsTmp)

				for _, eventName := range eventsToEnable {
					event, ok := events.FieldOk(eventName)
					if ok {
						event.Set(true)
						eventsEnabled = append(eventsEnabled, eventName)
					} else {
						eventsFailed = append(eventsFailed, eventName)
					}
				}
			}

			guildSettings[env.Guild.ID].LogSettings.LoggingEvents = *LoggingEventsTmp

			responseMessage := "Successfully enabled logging"
			if guildSettings[env.Guild.ID].LogSettings.LoggingChannel != "" {
				responseMessage += "."
			} else {
				responseMessage += " and set the logging channel to this channel."
			}
			if len(eventsToEnable) > 0 {
				responseMessage += "\n"
				if len(eventsEnabled) > 0 {
					responseMessage += "\nEnabled the following events: " + strings.Join(eventsEnabled, ", ")
				}
				if len(eventsFailed) > 0 {
					responseMessage += "\nFailed to find the following events: " + strings.Join(eventsFailed, ", ")
				}
			}
			return NewGenericEmbed("Server Settings - Log", responseMessage)
		case "disable":
			if len(args) == 3 && args[2] == "all" {
				guildSettings[env.Guild.ID].LogSettings.LoggingEvents = LogEvents{}
				return NewGenericEmbed("Server Settings - Log", "Successfully disabled all logging events.")
			}

			eventsToDisable := make([]string, 0)
			if len(args) > 2 {
				eventsToDisable = args[2:]
			}
			eventsDisabled := make([]string, 0)
			eventsFailed := make([]string, 0)

			if len(eventsToDisable) > 0 {
				events := structs.New(LoggingEventsTmp)

				for _, eventName := range eventsToDisable {
					event, ok := events.FieldOk(eventName)
					if ok {
						event.Set(false)
						eventsDisabled = append(eventsDisabled, eventName)
					} else {
						eventsFailed = append(eventsFailed, eventName)
					}
				}
			} else {
				guildSettings[env.Guild.ID].LogSettings.LoggingEnabled = false
				return NewGenericEmbed("Server Settings - Log", "Successfully disabled logging.")
			}

			guildSettings[env.Guild.ID].LogSettings.LoggingEvents = *LoggingEventsTmp

			responseMessage := ""
			if len(eventsToDisable) > 0 {
				if len(eventsDisabled) > 0 {
					responseMessage += "\nDisabled the following events: " + strings.Join(eventsDisabled, ", ")
				}
				if len(eventsFailed) > 0 {
					responseMessage += "\nFailed to find the following events: " + strings.Join(eventsFailed, ", ")
				}
			}
			return NewGenericEmbed("Server Settings - Log", responseMessage)
		case "events":
			responseMessage := "__Event states__\n"

			events := structs.New(guildSettings[env.Guild.ID].LogSettings.LoggingEvents)
			eventFields := events.Fields()

			for _, event := range eventFields {
				responseMessage += "\n" + event.Name() + ": **" + strconv.FormatBool(event.Value().(bool)) + "**"
			}

			return NewGenericEmbed("Server Settings - Log", responseMessage)
		default:
			return NewErrorEmbed("Server Settings - Log Error", "Unknown log command ``"+args[1]+"``.")
		}
	case "reset":
		if len(args) < 2 {
			return NewErrorEmbed("Server Settings - Reset Error", "You must specify a setting to reset.")
		}
		switch args[1] {
		case "joinmsg":
			guildSettings[env.Guild.ID].UserJoinMessage = ""
			guildSettings[env.Guild.ID].UserJoinMessageChannel = ""
		case "leavemsg":
			guildSettings[env.Guild.ID].UserLeaveMessage = ""
			guildSettings[env.Guild.ID].UserLeaveMessageChannel = ""
		case "log":
			guildSettings[env.Guild.ID].LogSettings.LoggingChannel = ""
			guildSettings[env.Guild.ID].LogSettings.LoggingEnabled = false
			guildSettings[env.Guild.ID].LogSettings.LoggingEvents = LogEvents{}
		default:
			return NewErrorEmbed("Server Settings - Reset Error", "Error finding the setting ``"+args[1]+"``.")
		}
		return NewGenericEmbed("Server Settings - Reset", "Successfully reset the settings for ``"+args[1]+"``.")
	default:
		return NewErrorEmbed("Server Settings Error", "Error finding the setting ``"+args[0]+"``.")
	}
	return nil
}
