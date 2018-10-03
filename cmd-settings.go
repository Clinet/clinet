package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fatih/structs"
)

func commandSettingsBot(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "prefix":
		if len(args) > 1 {
			if args[1] == botData.CommandPrefix {
				guildSettings[env.Guild.ID].BotPrefix = ""
			} else {
				guildSettings[env.Guild.ID].BotPrefix = args[1]
			}
			return NewGenericEmbed("Bot Settings - Command Prefix", "Successfully set the command prefix to ``"+strings.Replace(args[1], "`", "\\`", -1)+"``.")
		}
		if guildSettings[env.Guild.ID].BotPrefix != "" {
			return NewGenericEmbed("Bot Settings - Command Prefix", "Current command prefix:\n\n"+guildSettings[env.Guild.ID].BotPrefix)
		}
		return NewGenericEmbed("Bot Settings - Command Prefix", "Current command prefix:\n\n"+botData.CommandPrefix)
	}
	return NewErrorEmbed("Bot Settings Error", "Error finding the setting ``"+args[0]+"``.")
}

func commandSettingsUser(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//We're getting there (⟃ ͜ʖ ⟄)
	return nil
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
	case "filter":
		if len(args) < 2 {
			filterHelpCmd := &Command{
				HelpText: "Manages the swear filter for this server.",
				RequiredArguments: []string{
					"setting (value(s))",
				},
				Arguments: []CommandArgument{
					{Name: "enable", Description: "Enables the swear filter for this server", ArgType: "this"},
					{Name: "disable", Description: "Disables the swear filter for this server", ArgType: "this"},
					{Name: "timeout", Description: "Displays or sets the timeout for deleting warning messages", ArgType: "this/number"},
					{Name: "words", Description: "Lists filtered words, or adds/removes specified words/clears all words", ArgType: "this/(add word1)/(remove word2)/clear"},
				},
			}
			return getCustomCommandUsage(filterHelpCmd, "server filter", "Server Settings - Swear Filter Help")
		}

		switch args[1] {
		case "enable":
			guildSettings[env.Guild.ID].SwearFilter.Enabled = true
			return NewGenericEmbed("Server Settings - Swear Filter", "Successfully enabled the swear filter.")
		case "disable":
			guildSettings[env.Guild.ID].SwearFilter.Enabled = false
			return NewGenericEmbed("Server Settings - Swear Filter", "Successfully disabled the swear filter.")
		case "words":
			if len(args) < 3 {
				wordListEmbed := NewEmbed().
					SetTitle("Server Settings - Swear Filter").
					AddField("Filtered Words", strings.Join(guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords, ", ")).
					SetColor(0x1C1C1C).MessageEmbed
				return wordListEmbed
			}
			switch args[2] {
			case "add":
				if len(args) < 4 {
					return NewErrorEmbed("Server Settings - Swear Filter Error", "You must specify one or more words to add to the filter.")
				}
				guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords = append(guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords, args[3:]...)
				return NewGenericEmbed("Server Settings - Swear Filter", "Successfully added the provided words to the filter.")
			case "remove":
				if len(args) < 4 {
					return NewErrorEmbed("Server Settings - Swear Filter Error", "You must specify one or more words to remove from the filter.")
				}
				for _, word := range guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords {
					guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords = remove(guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords, word)
				}
				return NewGenericEmbed("Server Settings - Swear Filter", "Successfully removed the provided words from the filter.")
			case "clear":
				guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords = make([]string, 0)
				return NewGenericEmbed("Server Settings - Swear Filter", "Successfully cleared all words from the filter.")
			}
		}
	case "timeout":
		if len(args) < 3 {
			if guildSettings[env.Guild.ID].SwearFilter.WarningDeleteTimeout == 0 {
				return NewGenericEmbed("Server Settings - Swear Filter", "The timeout for deleting warning messages is disabled.")
			}
			timeout := strconv.Itoa(int(guildSettings[env.Guild.ID].SwearFilter.WarningDeleteTimeout))
			return NewGenericEmbed("Server Settings - Swear Filter", "The current timeout for deleting warning messages is set to "+timeout+" seconds.")
		}
		timeout, err := strconv.Atoi(args[2])
		if err != nil {
			return NewErrorEmbed("Server Settings - Swear Filter Error", "``"+args[2]+"`` is not a valid number.")
		}
		guildSettings[env.Guild.ID].SwearFilter.WarningDeleteTimeout = time.Duration(timeout)
		return NewGenericEmbed("Server Settings - Swear Filter", "Successfully set he timeout for deleting warning messages to "+args[2]+" seconds.")
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

		LoggingEventsTmp := &guildSettings[env.Guild.ID].LogSettings.LoggingEvents

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
		case "filter":
			guildSettings[env.Guild.ID].SwearFilter.Enabled = false
			guildSettings[env.Guild.ID].SwearFilter.BlacklistedWords = make([]string, 0)
			guildSettings[env.Guild.ID].SwearFilter.DisableNormalize = false
			guildSettings[env.Guild.ID].SwearFilter.DisableSpacedTab = false
			guildSettings[env.Guild.ID].SwearFilter.DisableMultiWhitespaceStripping = false
			guildSettings[env.Guild.ID].SwearFilter.DisableZeroWidthStripping = false
			guildSettings[env.Guild.ID].SwearFilter.DisableSpacedBypass = false
			guildSettings[env.Guild.ID].SwearFilter.WarningDeleteTimeout = time.Duration(0)
			guildSettings[env.Guild.ID].SwearFilter.AllowAdminBypass = false
			guildSettings[env.Guild.ID].SwearFilter.AllowBotOwnerBypass = false
		default:
			return NewErrorEmbed("Server Settings - Reset Error", "Error finding the setting ``"+args[1]+"``.")
		}
		return NewGenericEmbed("Server Settings - Reset", "Successfully reset the settings for ``"+args[1]+"``.")
	}
	return NewErrorEmbed("Server Settings Error", "Error finding the setting ``"+args[0]+"``.")
}
