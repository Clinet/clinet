package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Command holds data related to a command executable by any message system
type Command struct {
	Function            func([]string, *CommandEnvironment) *discordgo.MessageEmbed //The function value of what to execute when the command is ran
	HelpText            string                                                      //The text that will display in the help message
	Arguments           []CommandArgument                                           //The arguments required for this command
	RequiredArguments   []string                                                    //The minimum required arguments by name that must exist for the function to execute; default = 0
	RequiredPermissions int                                                         //The permission(s) a user must have for the command to be executed by them

	IsAlternateOf string //If this is an alternate command, point to the original command

	IsAdministrative bool //Whether or not this command requires the user to be a bot admin
}

// CommandArgument holds data related to an argument available or required by a command
type CommandArgument struct {
	Name        string //The name of the argument
	ArgType     string //The argument's type
	Description string //A description of the argument
}

// CommandEnvironment holds data related to the environment a command can utilize for data or functionality
type CommandEnvironment struct {
	Channel *discordgo.Channel //The channel the command was executed in
	Guild   *discordgo.Guild   //The guild the command was executed in
	Message *discordgo.Message //The message that triggered the command execution
	User    *discordgo.User    //The user that executed the command

	Command string //The command used to execute the command with this environment (in the event of a command alias)

	UpdatedMessageEvent bool
}

func initCommands() {
	//Initialize the commands map
	botData.Commands = make(map[string]*Command)

	//All user-accessible commands with no parameters
	botData.Commands["about"] = &Command{Function: commandAbout, HelpText: "Displays information about " + botData.BotName + " and how to use it."}
	botData.Commands["invite"] = &Command{Function: commandInvite, HelpText: "Displays available invite links for " + botData.BotName + "."}
	botData.Commands["donate"] = &Command{Function: commandDonate, HelpText: "Displays available donation links for " + botData.BotName + "."}
	botData.Commands["source"] = &Command{Function: commandSource, HelpText: "Displays available source code links for " + botData.BotName + "."}
	botData.Commands["version"] = &Command{Function: commandVersion, HelpText: "Displays the current version of " + botData.BotName + "."}
	botData.Commands["credits"] = &Command{Function: commandCredits, HelpText: "Displays a list of credits for the creation and functionality of " + botData.BotName + "."}
	botData.Commands["roll"] = &Command{Function: commandRoll, HelpText: "Rolls a dice."}
	botData.Commands["doubleroll"] = &Command{Function: commandDoubleRoll, HelpText: "Rolls two die."}
	botData.Commands["coinflip"] = &Command{Function: commandCoinFlip, HelpText: "Flips a coin."}
	botData.Commands["join"] = &Command{Function: commandVoiceJoin, HelpText: "Joins the current voice channel.", RequiredPermissions: discordgo.PermissionVoiceConnect}
	botData.Commands["leave"] = &Command{Function: commandVoiceLeave, HelpText: "Leaves the current voice channel.", RequiredPermissions: discordgo.PermissionVoiceConnect}
	botData.Commands["ping"] = &Command{Function: commandPing, HelpText: "Returns the ping average to Discord."}

	//All user-accessible commands with parameters
	botData.Commands["help"] = &Command{
		Function: commandHelp,
		HelpText: "Displays a list of commands you have permission to use.",
		Arguments: []CommandArgument{
			{Name: "page", Description: "The help page to view", ArgType: "number"},
			{Name: "command", Description: "The command to view help for", ArgType: "string"},
		},
	}
	botData.Commands["hewwo"] = &Command{
		Function: commandHewwo,
		HelpText: "Hewwo!!! (´・ω・｀)",
		RequiredArguments: []string{
			"message",
		},
		Arguments: []CommandArgument{
			{Name: "message", Description: "The text to translate to Hewwo", ArgType: "string"},
		},
	}
	botData.Commands["zalgo"] = &Command{
		Function: commandZalgo,
		HelpText: "Mystifies your text.",
		RequiredArguments: []string{
			"message",
		},
		Arguments: []CommandArgument{
			{Name: "message", Description: "The text to mystify", ArgType: "string"},
		},
	}
	botData.Commands["nlp"] = &Command{
		Function: commandNLP,
		HelpText: "Raw natural language processing in Discord. Powered by Prose:tm:.",
		RequiredArguments: []string{
			"message",
		},
		Arguments: []CommandArgument{
			{Name: "message", Description: "The message to parse", ArgType: "string"},
		},
	}
	botData.Commands["image"] = &Command{
		Function:            commandImage,
		HelpText:            "Allows you to manipulate images with various filters and encodings.",
		RequiredPermissions: discordgo.PermissionAttachFiles,
		RequiredArguments: []string{
			"effect",
		},
		Arguments: []CommandArgument{
			{Name: "fliphorizontal", Description: "Flips the image horizontally", ArgType: "this"},
			{Name: "flipvertical", Description: "Flips the image vertically", ArgType: "this"},
			{Name: "grayscale/greyscale", Description: "Applies a grayscale effect to the image", ArgType: "this"},
			{Name: "invert", Description: "Inverts the colors of the image", ArgType: "this"},
			{Name: "rotate90", Description: "Rotates the image by 90° clockwise", ArgType: "this"},
			{Name: "rotate180", Description: "Rotates the image by 180° clockwise", ArgType: "this"},
			{Name: "rotate270", Description: "Rotates the image by 270° clockwise", ArgType: "this"},
			{Name: "sobel", Description: "Applies the Sobel filter to the image", ArgType: "this"},
			{Name: "transpose", Description: "Transposes the image", ArgType: "this"},
			{Name: "transverse", Description: "Transverses the image", ArgType: "this"},
			{Name: "test", Description: "Applies a testing convolution effect to the image", ArgType: "this"},
		},
	}
	botData.Commands["cve"] = &Command{
		Function: commandCVE,
		HelpText: "Fetches information about a specified CVE.",
		RequiredArguments: []string{
			"CVE ID",
		},
		Arguments: []CommandArgument{
			{Name: "cve", Description: "The CVE ID to fetch information about", ArgType: "string"},
		},
	}
	botData.Commands["geoip"] = &Command{
		Function: commandGeoIP,
		HelpText: "Performs a GeoIP lookup on the specified IP/hostname.",
		RequiredArguments: []string{
			"IP/hostname",
		},
		Arguments: []CommandArgument{
			{Name: "IP/hostname", Description: "The IP or hostname to perform a GeoIP lookup on", ArgType: "IP address/hostname"},
		},
	}
	if botData.BotOptions.UseXKCD {
		botData.Commands["xkcd"] = &Command{
			Function: commandXKCD,
			HelpText: "Displays an XKCD comic depending on the requested type or comic number.",
			RequiredArguments: []string{
				"(comic number|latest|random)",
			},
			Arguments: []CommandArgument{
				{Name: "comic number", Description: "The number of an existing XKCD comic", ArgType: "number"},
				{Name: "latest", Description: "Fetches the latest XKCD comic", ArgType: "this"},
				{Name: "random", Description: "Fetches a random XKCD comic", ArgType: "this"},
			},
		}
	}
	if botData.BotOptions.UseImgur {
		botData.Commands["imgur"] = &Command{
			Function: commandImgur,
			HelpText: "Displays info about the specified Imgur image or album URL.",
			RequiredArguments: []string{
				"url",
			},
			Arguments: []CommandArgument{
				{Name: "url", Description: "The Imgur image or album URL", ArgType: "string"},
			},
		}
	}
	if botData.BotOptions.UseGitHub {
		botData.Commands["github"] = &Command{
			Function: commandGitHub,
			HelpText: "Displays info about the specified GitHub user or repo.",
			RequiredArguments: []string{
				"username(/repo)",
			},
			Arguments: []CommandArgument{
				{Name: "username", Description: "The GitHub user to fetch info about", ArgType: "string"},
				{Name: "username/repo", Description: "The GitHub repo to fetch info about", ArgType: "string"},
			},
		}
	}
	botData.Commands["urbandictionary"] = &Command{
		Function: commandUrbanDictionary,
		HelpText: "Displays the definition of a term according to the Urban Dictionary.",
		RequiredArguments: []string{
			"term",
		},
		Arguments: []CommandArgument{
			{Name: "term", Description: "The term to fetch a definition for", ArgType: "string"},
		},
	}

	//Voice commands
	botData.Commands["play"] = &Command{
		Function:            commandPlay,
		HelpText:            "Plays either the first result from a YouTube search query or the specified stream URL in the user's voice channel.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
		Arguments: []CommandArgument{
			{Name: "search query", Description: "The YouTube search query to use when fetching a video to play", ArgType: "string"},
			{Name: "url", Description: "The YouTube, SoundCloud, or direct audio/video URL to play", ArgType: "string"},
		},
	}
	botData.Commands["stop"] = &Command{
		Function:            commandStop,
		HelpText:            "Stops the audio playback in the user's voice channel.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}
	botData.Commands["skip"] = &Command{
		Function:            commandSkip,
		HelpText:            "Skips to the next queue entry in the user's voice channel.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}
	botData.Commands["pause"] = &Command{
		Function:            commandPause,
		HelpText:            "Pauses the audio playback in the user's voice channel.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}
	botData.Commands["resume"] = &Command{
		Function:            commandResume,
		HelpText:            "Resumes the audio playback in the user's voice channel.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}
	botData.Commands["volume"] = &Command{
		Function:            commandVolume,
		HelpText:            "Sets the volume level for the next audio playback.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
		RequiredArguments: []string{
			"volume",
		},
		Arguments: []CommandArgument{
			{Name: "volume", Description: "The volume level to use", ArgType: "number [0 - 512]"},
		},
	}
	botData.Commands["repeat"] = &Command{
		Function:            commandRepeat,
		HelpText:            "Switches queue playback between three modes: no repeat, repeat queue, and repeat now playing.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}
	botData.Commands["shuffle"] = &Command{
		Function:            commandShuffle,
		HelpText:            "Toggles queue shuffling during playback.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}
	botData.Commands["youtube"] = &Command{
		Function:            commandYouTube,
		HelpText:            "Allows you to navigate YouTube search results to select what to add to the queue.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
		RequiredArguments: []string{
			"command (value)",
		},
		Arguments: []CommandArgument{
			{Name: "search", Description: "Searches the specified query value", ArgType: "string"},
			{Name: "next", Description: "Navigates forward in a search result's pages", ArgType: "this"},
			{Name: "previous", Description: "Navigates backward in a search result's pages", ArgType: "this"},
			{Name: "cancel", Description: "Cancels the search result", ArgType: "this"},
			{Name: "play", Description: "Plays the chosen search result from the current page", ArgType: "number"},
		},
	}
	botData.Commands["queue"] = &Command{
		Function:            commandQueue,
		HelpText:            "Lists and manages entries in the queue.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
		Arguments: []CommandArgument{
			{Name: "clear", Description: "Clears the queue", ArgType: "this"},
			{Name: "remove", Description: "Removes the specified queue entry or entries", ArgType: "number"},
		},
	}
	botData.Commands["nowplaying"] = &Command{
		Function:            commandNowPlaying,
		HelpText:            "Displays the now playing entry.",
		RequiredPermissions: discordgo.PermissionVoiceConnect,
	}

	//All moderation commands with parameters
	botData.Commands["purge"] = &Command{
		Function:            commandPurge,
		HelpText:            "Purges the specified amount of messages from the channel, up to 100 messages at a time.",
		RequiredPermissions: discordgo.PermissionManageMessages,
		RequiredArguments: []string{
			"amount (user1) (user2) (user3)",
		},
		Arguments: []CommandArgument{
			{Name: "message count", Description: "The amount of messages to delete", ArgType: "number"},
			{Name: "user(s)", Description: "The user(s) to delete the messages from within the specified amount of messages", ArgType: "mention"},
		},
	}
	botData.Commands["kick"] = &Command{
		Function:            commandKick,
		HelpText:            "Kicks the specified user(s) from the server.",
		RequiredPermissions: discordgo.PermissionKickMembers,
		RequiredArguments: []string{
			"user1 (user2) (user3) (reason for kick)",
		},
		Arguments: []CommandArgument{
			{Name: "user(s)", Description: "The user(s) to kick", ArgType: "mention"},
			{Name: "reason", Description: "The reason for the kick", ArgType: "string"},
		},
	}
	botData.Commands["ban"] = &Command{
		Function:            commandBan,
		HelpText:            "Bans the specified user(s) from the server.",
		RequiredPermissions: discordgo.PermissionBanMembers,
		RequiredArguments: []string{
			"(days) user1 (user2) (user3) (reason for ban)",
		},
		Arguments: []CommandArgument{
			{Name: "days", Description: "How many days worth of messages to delete from the specified user(s)", ArgType: "number"},
			{Name: "user(s)", Description: "The user(s) to ban", ArgType: "mention"},
			{Name: "reason", Description: "The reason for the ban", ArgType: "string"},
		},
	}

	botData.Commands["server"] = &Command{
		Function:            commandSettingsServer,
		HelpText:            "Changes the specified settings for the server.",
		RequiredPermissions: discordgo.PermissionAdministrator,
		RequiredArguments: []string{
			"setting (value)",
		},
		Arguments: []CommandArgument{
			{Name: "filter", Description: "Manages the swear filter", ArgType: "this"},
			{Name: "joinmsg", Description: "Sets the join message for this channel", ArgType: "string"},
			{Name: "leavemsg", Description: "Sets the leave message for this channel", ArgType: "string"},
			{Name: "log", Description: "Manages the logging events", ArgType: "this"},
			{Name: "reset", Description: "Resets the specified setting to the default/empty value", ArgType: "string"},
		},
	}
	botData.Commands["bot"] = &Command{
		Function:            commandSettingsBot,
		HelpText:            "Changes the specified settings for the bot within this server.",
		RequiredPermissions: discordgo.PermissionAdministrator,
		RequiredArguments: []string{
			"setting (value)",
		},
		Arguments: []CommandArgument{
			{Name: "prefix", Description: "Sets the bot command prefix", ArgType: "string"},
		},
	}

	botData.Commands["starboard"] = &Command{
		Function:            commandStarboard,
		HelpText:            "Manages the guild's starboard.",
		RequiredPermissions: discordgo.PermissionAdministrator,
		RequiredArguments: []string{
			"setting (value)",
		},
		Arguments: []CommandArgument{
			{Name: "enable", Description: "Enables the starboard", ArgType: "this"},
			{Name: "disable", Description: "Disables the starboard", ArgType: "this"},
			{Name: "channel (set/remove)", Description: "Either returns the current starboard channel or optionally sets it to the current channel or removes the current channel in place", ArgType: "this"},
			{Name: "nsfwchannel (set/remove)", Description: "Either returns the current NSFW channel or optionally sets it to the current channel (if marked as NSFW) or removes the current channel in place", ArgType: "this"},
			{Name: "emoji (emoji)", Description: "Either returns the current emoji or optionally sets it to the specified emoji", ArgType: "emoji"},
			{Name: "nsfwemoji (emoji)", Description: "Either returns the current NSFW emoji or optionally sets it to the specified emoji", ArgType: "emoji"},
			{Name: "selfstar", Description: "Sets whether or not selfstars are permitted", ArgType: "boolean"},
			{Name: "minimum", Description: "Either returns the current minimum reaction requirement or sets it to the specified amount", ArgType: "number"},
		},
	}

	/*
		botData.Commands["user"] = &Command{
			Function: commandSettingsUser,
			HelpText: "Changes settings pertaining to you.",
			RequiredArguments: []string{
				"setting",
				"(value)",
			},
			Arguments: []CommandArgument{
				{Name: "description", Description: "Sets your user description", ArgType: "string"},
			},
		}
	*/

	//Alternate commands for pre-established commands
	botData.Commands["?"] = &Command{IsAlternateOf: "help"}
	botData.Commands["commands"] = &Command{IsAlternateOf: "help"}
	botData.Commands["ver"] = &Command{IsAlternateOf: "version"}
	botData.Commands["v"] = &Command{IsAlternateOf: "version"}
	botData.Commands["rolldouble"] = &Command{IsAlternateOf: "doubleroll"}
	botData.Commands["flipcoin"] = &Command{IsAlternateOf: "coinflip"}
	botData.Commands["img"] = &Command{IsAlternateOf: "image"}
	botData.Commands["gh"] = &Command{IsAlternateOf: "github"}
	botData.Commands["yt"] = &Command{IsAlternateOf: "youtube"}
	botData.Commands["np"] = &Command{IsAlternateOf: "nowplaying"}
	botData.Commands["ud"] = &Command{IsAlternateOf: "urbandictionary"}
	botData.Commands["owo"] = &Command{IsAlternateOf: "hewwo"}
	botData.Commands["uwu"] = &Command{IsAlternateOf: "uwu"}

	//Testing commands, only available if debug mode is enabled
	if botData.DebugMode {
		botData.Commands["botinfo"] = &Command{
			Function: commandBotInfo,
			HelpText: "Displays info about the bot's current state.",
		}
	}

	//Administrative commands for bot owners
	botData.Commands["reload"] = &Command{Function: commandReload, HelpText: "Reloads the bot configuration.", IsAdministrative: true}
	botData.Commands["restart"] = &Command{Function: commandRestart, HelpText: "Restarts the bot in case something goes awry.", IsAdministrative: true}
	botData.Commands["update"] = &Command{Function: commandUpdate, HelpText: "Updates the bot to the latest git repo commit.", IsAdministrative: true}
}

func callCommand(commandName string, args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if command, exists := botData.Commands[commandName]; exists {
		if command.IsAlternateOf != "" {
			if commandAlternate, exists := botData.Commands[command.IsAlternateOf]; exists {
				command = commandAlternate
			} else {
				return nil
			}
		}
		if command.IsAdministrative && env.User.ID != botData.BotOwnerID {
			return NewErrorEmbed("Command Error - Not Authorized (NA)", "You are not authorized to use this command.")
		}
		if permissionsAllowed, _ := MemberHasPermission(botData.DiscordSession, env.Guild.ID, env.User.ID, env.Channel.ID, discordgo.PermissionAdministrator|command.RequiredPermissions); permissionsAllowed || command.RequiredPermissions == 0 {
			if len(args) >= len(command.RequiredArguments) {
				return command.Function(args, env)
			}
			return getCommandUsage(commandName, "Command Error - Not Enough Parameters (NEP)")
		}
		return NewErrorEmbed("Command Error - No Permissions (NP)", "You do not have the necessary permissions to use this command.")
	}
	return nil
}

func getCommandUsage(commandName, title string) *discordgo.MessageEmbed {
	command := botData.Commands[commandName]
	if command.IsAlternateOf != "" {
		command = botData.Commands[command.IsAlternateOf]
	}

	parameterFields := []*discordgo.MessageEmbedField{}
	parameterFields = append(parameterFields, &discordgo.MessageEmbedField{Name: "Usage", Value: botData.CommandPrefix + commandName + " " + strings.Join(command.RequiredArguments, " ")})
	for i := 0; i < len(command.Arguments); i++ {
		parameterFields = append(parameterFields, &discordgo.MessageEmbedField{Name: command.Arguments[i].Name + " (" + command.Arguments[i].ArgType + ")", Value: command.Arguments[i].Description, Inline: true})
	}

	usageEmbed := NewEmbed().
		SetTitle(title).
		SetDescription("**" + commandName + "**: " + command.HelpText).
		SetColor(0xFF0000).MessageEmbed
	usageEmbed.Fields = parameterFields

	return usageEmbed
}

func getCustomCommandUsage(command *Command, commandName, title string) *discordgo.MessageEmbed {
	parameterFields := []*discordgo.MessageEmbedField{}
	parameterFields = append(parameterFields, &discordgo.MessageEmbedField{Name: "Usage", Value: botData.CommandPrefix + commandName + " " + strings.Join(command.RequiredArguments, " ")})
	for i := 0; i < len(command.Arguments); i++ {
		parameterFields = append(parameterFields, &discordgo.MessageEmbedField{Name: command.Arguments[i].Name + " (" + command.Arguments[i].ArgType + ")", Value: command.Arguments[i].Description, Inline: true})
	}

	usageEmbed := NewEmbed().
		SetTitle(title).
		SetDescription("**" + commandName + "**: " + command.HelpText).
		SetColor(0xFF0000).MessageEmbed
	usageEmbed.Fields = parameterFields

	return usageEmbed
}
