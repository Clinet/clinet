package main

import (
	"bytes"
	"image"
	"image/png"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/JoshuaDoes/go-cve"
	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/gift"
	"github.com/rylio/ytdl"
)

// Command holds data related to a command executable by any message system
type Command struct {
	Function          func([]string, *CommandEnvironment) *discordgo.MessageEmbed //The function value of what to execute when the command is ran
	HelpText          string                                                      //The text that will display in the help message
	Arguments         []CommandArgument                                           //The arguments required for this command
	RequiredArguments []string                                                    //The minimum required arguments by name that must exist for the function to execute; default = 0

	IsAlternateOf       string //If this is an alternate command, point to the original command
	RequiredPermissions int    //The permission(s) a user must have for the command to be executed by them
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
	botData.Commands["help"] = &Command{Function: commandHelp, HelpText: "Displays a list of commands you have permission to use."}
	botData.Commands["about"] = &Command{Function: commandAbout, HelpText: "Displays information about " + botData.BotName + " and how to use it."}
	botData.Commands["version"] = &Command{Function: commandVersion, HelpText: "Displays the current version of " + botData.BotName + "."}
	botData.Commands["credits"] = &Command{Function: commandCredits, HelpText: "Displays a list of credits for the creation and functionality of " + botData.BotName + "."}
	botData.Commands["roll"] = &Command{Function: commandRoll, HelpText: "Rolls a dice."}
	botData.Commands["doubleroll"] = &Command{Function: commandDoubleRoll, HelpText: "Rolls two die."}
	botData.Commands["coinflip"] = &Command{Function: commandCoinFlip, HelpText: "Flips a coin."}
	botData.Commands["join"] = &Command{Function: commandVoiceJoin, HelpText: "Joins the current voice channel.", RequiredPermissions: discordgo.PermissionVoiceConnect}
	botData.Commands["leave"] = &Command{Function: commandVoiceLeave, HelpText: "Leaves the current voice channel.", RequiredPermissions: discordgo.PermissionVoiceConnect}

	//All user-accessible commands with parameters
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
	//Pause, resume, repeat, shuffle
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
			{Name: "select", Description: "Selects the chosen search result from the current page", ArgType: "number"},
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
			"setting",
			"(value)",
		},
		Arguments: []CommandArgument{
			{Name: "joinmsg", Description: "Sets the join message for this channel", ArgType: "string"},
			{Name: "leavemsg", Description: "Sets the leave message for this channel", ArgType: "string"},
			{Name: "reset", Description: "Resets the specified setting to the default/empty value", ArgType: "string"},
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

	//Testing commands, only available if debug mode is enabled
	if botData.DebugMode {
		botData.Commands["botinfo"] = &Command{
			Function: commandBotInfo,
			HelpText: "Displays info about the bot's current state.",
		}
	}
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
		if permissionsAllowed, _ := MemberHasPermission(botData.DiscordSession, env.Guild.ID, env.User.ID, discordgo.PermissionAdministrator|command.RequiredPermissions); permissionsAllowed || command.RequiredPermissions == 0 {
			if len(args) >= len(command.RequiredArguments) {
				return command.Function(args, env)
			}
			return getCommandUsage(commandName, "Command Error - Not Enough Parameters (NEP)")
		}
		return NewErrorEmbed("Command Error - No Permissions (NP)", "You do not have the necessary permissions to use this command.")
	}
	return botData.Commands[commandName].Function(args, env)
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

func commandHelp(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//First see if help text is being requested for a particular command
	if len(args) > 0 {
		if command, exists := botData.Commands[args[0]]; exists {
			if command.IsAlternateOf != "" {
				if commandAlternate, exists := botData.Commands[command.IsAlternateOf]; exists {
					command = commandAlternate
				} else {
					return nil
				}
			}
			return getCommandUsage(args[0], "Help for **"+args[0]+"**")
		}
	}

	//Before we fetch help text data, we need to have an alphabetical listing of commands
	var commandMapKeys []string
	for commandMapKey := range botData.Commands {
		commandMapKeys = append(commandMapKeys, commandMapKey)
	}
	sort.Strings(commandMapKeys)

	//Create a dynamic list of fields for the help embed
	commandFields := []*discordgo.MessageEmbedField{}

	//Iterate over the alphabetically sorted command list and add each listed command to the help embed field list
	for _, commandName := range commandMapKeys {
		command := botData.Commands[commandName]
		if command.IsAlternateOf == "" {
			if permissionsAllowed, _ := MemberHasPermission(botData.DiscordSession, env.Guild.ID, env.User.ID, discordgo.PermissionAdministrator|command.RequiredPermissions); permissionsAllowed || command.RequiredPermissions == 0 {
				commandField := &discordgo.MessageEmbedField{Name: botData.CommandPrefix + commandName, Value: command.HelpText, Inline: true}
				commandFields = append(commandFields, commandField)
			}
		}
	}

	//Create the help embed and give it the command list
	helpEmbed := NewEmbed().
		SetTitle(botData.BotName + " - Help").
		SetDescription("A list of commands you have permission to use.").
		SetColor(0xFAFAFA).MessageEmbed
	helpEmbed.Fields = commandFields

	//Return the help embed to the caller
	return helpEmbed
}
func commandAbout(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - About").
		SetDescription(botData.BotName+" is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
		AddField("How can I use "+botData.BotName+" in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
		AddField("How can I help keep "+botData.BotName+" running?", "The best ways to help keep "+botData.BotName+" running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
		AddField("How can I use "+botData.BotName+"?", "There are many ways to make use of "+botData.BotName+".\n1) Type ``cli$help`` and try using some of the available commands.\n2) Ask "+botData.BotName+" a question, ex: ``@"+botData.BotName+"#1823, what time is it?`` or ``@"+botData.BotName+"#1823, what is DiscordApp?``.").
		AddField("Where can I join the "+botData.BotName+" Discord server?", "If you would like to get help and support with "+botData.BotName+" or experiment with the latest and greatest of "+botData.BotName+", use the Discord Server Invite Link at the end of this message.").
		AddField("Invite Link", "https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot").
		AddField("Donation Link", "https://www.paypal.me/JoshuaDoes").
		AddField("Source Code Link", "https://github.com/JoshuaDoes/clinet-discord/").
		AddField("Discord Server Invite Link", "https://discord.gg/qkbKEWT").
		SetColor(0x1C1C1C).MessageEmbed
}
func commandVersion(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Version").
		AddField("Build ID", BuildID).
		AddField("Build Date", BuildDate).
		AddField("Latest Development", GitCommitMsg).
		AddField("GitHub Commit URL", GitHubCommitURL).
		AddField("Golang Version", GolangVersion).
		SetColor(0x1C1C1C).MessageEmbed
}
func commandCredits(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Credits").
		AddField("Bot Development", "- JoshuaDoes (2018)").
		AddField("Programming Language", "- Golang").
		AddField("Golang Libraries", "- https://github.com/bwmarrin/discordgo\n"+
			"- https://github.com/disintegration/gift\n"+
			"- https://github.com/JoshuaDoes/duckduckgolang\n"+
			"- https://github.com/google/go-github/github\n"+
			"- https://github.com/jonas747/dca\n"+
			"- https://github.com/JoshuaDoes/go-soundcloud\n"+
			"- https://github.com/JoshuaDoes/go-wolfram\n"+
			"- https://github.com/koffeinsource/go-imgur\n"+
			"- https://github.com/koffeinsource/go-klogger\n"+
			"- https://github.com/nishanths/go-xkcd\n"+
			"- https://github.com/paked/configure\n"+
			"- https://github.com/robfig/cron\n"+
			"- https://github.com/rylio/ytdl\n"+
			"- https://google.golang.org/api/googleapi/transport\n"+
			"- https://google.golang.org/api/youtube/v3").
		AddField("Icon Design", "- thejsa").
		AddField("Source Code", "- https://github.com/JoshuaDoes/clinet-discord").
		SetColor(0x1C1C1C).MessageEmbed
}
func commandRoll(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	random := rand.Intn(6) + 1
	return NewGenericEmbed("Roll", "You rolled a "+strconv.Itoa(random)+"!")
}
func commandDoubleRoll(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	random1 := rand.Intn(6) + 1
	random2 := rand.Intn(6) + 1
	randomTotal := random1 + random2
	return NewGenericEmbed("Double Roll", "You rolled a "+strconv.Itoa(random1)+" and a "+strconv.Itoa(random2)+". The total is "+strconv.Itoa(randomTotal)+"!")
}
func commandCoinFlip(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	random := rand.Intn(2)
	switch random {
	case 0:
		return NewGenericEmbed("Coin Flip", "The coin landed on heads!")
	case 1:
		return NewGenericEmbed("Coin Flip", "The coin landed on tails!")
	}
	return nil
}
func commandImage(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(args) > 1 {
		if len(env.Message.Attachments) > 0 {
			for _, attachment := range env.Message.Attachments {
				srcImageURL := attachment.URL
				srcImageHTTP, err := http.Get(srcImageURL)
				if err != nil {
					return NewErrorEmbed("Image Error", "Unable to fetch image.")
				}
				srcImage, _, err := image.Decode(srcImageHTTP.Body)
				if err != nil {
					return NewErrorEmbed("Image Error", "Unable to decode image.")
				}

				g := &gift.GIFT{}
				var outImage bytes.Buffer

				switch args[0] {
				case "fliphorizontal":
					g = gift.New(gift.FlipHorizontal())
				case "flipvertical":
					g = gift.New(gift.FlipVertical())
				case "grayscale", "greyscale":
					g = gift.New(gift.Grayscale())
				case "invert":
					g = gift.New(gift.Invert())
				case "rotate90":
					g = gift.New(gift.Rotate90())
				case "rotate180":
					g = gift.New(gift.Rotate180())
				case "rotate270":
					g = gift.New(gift.Rotate270())
				case "sobel":
					g = gift.New(gift.Sobel())
				case "transpose":
					g = gift.New(gift.Transpose())
				case "transverse":
					g = gift.New(gift.Transverse())
				case "test":
					g = gift.New(gift.Convolution(
						[]float32{
							-1, -1, 0,
							-1, 1, 1,
							0, 1, 1,
						},
						false, false, false, 0.0,
					))
				}

				dstImage := image.NewRGBA(g.Bounds(srcImage.Bounds()))
				g.Draw(dstImage, srcImage)

				err = png.Encode(&outImage, dstImage)
				if err != nil {
					return NewErrorEmbed("Image Error", "Unable to encode processed image.")
				}
				botData.DiscordSession.ChannelMessageSendComplex(env.Channel.ID, &discordgo.MessageSend{
					Content: "Processed image:",
					File: &discordgo.File{
						Name:   args[0] + ".png",
						Reader: &outImage,
					},
					Embed: &discordgo.MessageEmbed{
						Image: &discordgo.MessageEmbedImage{
							URL: "attachment://" + args[0] + ".png",
						},
					},
				})
			}
		} else {
			return NewErrorEmbed("Image Error", "You must upload an image to process.")
		}
	}
	return nil
}

func commandCVE(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	cveData, err := cve.GetCVE(args[0])
	if err != nil {
		return NewErrorEmbed("CVE Error", "There was an error fetching information about CVE ``"+args[0]+"``.")
	}
	return NewEmbed().
		SetTitle(args[0]).
		SetDescription(cveData.Summary).
		SetColor(0xC93130).MessageEmbed
}

func commandXKCD(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "latest":
		comic, err := botData.BotClients.XKCD.Latest()
		if err != nil {
			return NewErrorEmbed("XKCD Error", "There was an error fetching the latest XKCD comic.")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	case "random":
		comic, err := botData.BotClients.XKCD.Random()
		if err != nil {
			return NewErrorEmbed("XKCD Error", "There was an error fetching a random XKCD comic.")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	default:
		comicNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return NewErrorEmbed("XKCD Error", "``"+args[0]+"`` is not a valid number.")
		}

		comic, err := botData.BotClients.XKCD.Get(comicNumber)
		if err != nil {
			return NewErrorEmbed("XKCD Error", "There was an error fetching XKCD comic #"+args[0]+".")
		}
		return NewEmbed().
			SetTitle("xkcd - #" + args[0]).
			SetDescription(comic.Title).
			SetImage(comic.ImageURL).
			SetColor(0x96A8C8).MessageEmbed
	}
}

func commandImgur(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	responseEmbed, err := queryImgur(args[0])
	if err != nil {
		return NewErrorEmbed("Imgur Error", "There was an error fetching information about the specified URL.")
	}
	return responseEmbed
}

func commandGitHub(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	request := strings.Split(args[0], "/")

	switch len(request) {
	case 1: //Only user was specified
		user, err := GitHubFetchUser(request[0])
		if err != nil {
			return NewErrorEmbed("GitHub Error", "There was an error fetching information about the specified user.")
		}

		fields := []*discordgo.MessageEmbedField{}

		//Gather user info
		if user.Bio != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Bio", Value: *user.Bio})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Username", Value: *user.Login})
		if user.Name != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Name", Value: *user.Name})
		}
		if user.Company != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Company", Value: *user.Company})
		}
		if *user.Blog != "" {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Blog", Value: *user.Blog})
		}
		if user.Location != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Location", Value: *user.Location})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Public Repos", Value: strconv.Itoa(*user.PublicRepos)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Public Gists", Value: strconv.Itoa(*user.PublicGists)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Following", Value: strconv.Itoa(*user.Following)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Followers", Value: strconv.Itoa(*user.Followers)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "GitHub URL", Value: *user.HTMLURL})

		//Build embed about user
		responseEmbed := NewEmbed().
			SetTitle("GitHub User: " + *user.Login).
			SetImage(*user.AvatarURL).
			SetColor(0x24292D).MessageEmbed
		responseEmbed.Fields = fields

		return responseEmbed
	case 2: //Repo was specified
		repo, err := GitHubFetchRepo(request[0], request[1])
		if err != nil {
			return NewErrorEmbed("GitHub Error", "There was an error fetchign information about the specified repo.")
		}

		fields := []*discordgo.MessageEmbedField{}

		//Gather repo info
		if repo.Description != nil && *repo.Description != "" {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Description", Value: *repo.Description})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Name", Value: *repo.FullName})
		if repo.Homepage != nil && *repo.Homepage != "" {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Homepage", Value: *repo.Homepage})
		}
		if len(repo.Topics) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Topics", Value: strings.Join(repo.Topics, ", ")})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Default Branch", Value: *repo.DefaultBranch})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Is Fork", Value: strconv.FormatBool(*repo.Fork)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Forks", Value: strconv.Itoa(*repo.ForksCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Networks", Value: strconv.Itoa(*repo.NetworkCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Open Issues", Value: strconv.Itoa(*repo.OpenIssuesCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Stargazers", Value: strconv.Itoa(*repo.StargazersCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Subscribers", Value: strconv.Itoa(*repo.SubscribersCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Watchers", Value: strconv.Itoa(*repo.WatchersCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "GitHub URL", Value: *repo.HTMLURL})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Clone URL", Value: *repo.CloneURL})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Git URL", Value: *repo.GitURL})

		//Build embed about repo
		responseEmbed := NewEmbed().
			SetTitle("GitHub Repo: " + *repo.FullName).
			SetColor(0x24292D).MessageEmbed
		responseEmbed.Fields = fields

		return responseEmbed
	}

	return NewErrorEmbed("GitHub Error", "You must specify a GitHub user or a GitHub repo to fetch info about.\n\nExamples:\n```"+botData.CommandPrefix+"github JoshuaDoes\n"+botData.CommandPrefix+"gh JoshuaDoes/clinet-discord```")
}

func commandVoiceJoin(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
			return NewGenericEmbed("Voice", "Joined the voice channel.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel to use before using the join command.")
}

func commandVoiceLeave(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			if voiceIsStreaming(env.Guild.ID) {
				voiceStop(env.Guild.ID)
			}
			err := voiceLeave(env.Guild.ID, voiceState.ChannelID)
			if err != nil {
				return NewErrorEmbed("Vocie Error", "There was an error leaving the voice channel.")
			}
			return NewGenericEmbed("Voice", "Left the voice channel.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel to use before using the leave command.")
}

func commandPlay(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if env.UpdatedMessageEvent {
		//Todo: Remove this once I figure out how to detect if message update was user-triggered
		//Reason: If you use a YouTube/SoundCloud URL, Discord automatically updates the message with an embed
		//As far as I know, bots have no way to know if this was a Discord- or user-triggered message update
		//I eventually want users to be able to edit their play command to change a now playing or a queue entry that was misspelled
		return nil
	}

	for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
		//Wait for the handling of a previous playback command to finish
	}
	foundVoiceChannel := false
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			foundVoiceChannel = true
			voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
			break
		}
	}
	if !foundVoiceChannel {
		return NewErrorEmbed("Voice Error", "You must join the voice channel to use before using the play command.")
	}
	//Prevent other play commands in this voice session from messing up this process
	guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = true

	if len(args) >= 1 { //Query or URL was specified
		_, err := url.ParseRequestURI(args[0]) //Check to see if the first parameter is a URL
		if err != nil {                        //First parameter is not a URL
			queryURL, err := voiceGetQuery(strings.Join(args, " "))
			if err != nil {
				guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
				return NewErrorEmbed("Voice Error", "There was an error getting a result for the specified query.")
			}
			queueData := AudioQueueEntry{MediaURL: queryURL, Requester: env.Message.Author, Type: "youtube"}
			queueData.FillMetadata()
			if voiceIsStreaming(env.Guild.ID) {
				guildData[env.Guild.ID].QueueAdd(queueData)
				guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
				return queueData.GetQueueAddedEmbed()
			}
			guildData[env.Guild.ID].AudioNowPlaying = queueData
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)
			return queueData.GetNowPlayingEmbed()
		}

		//First parameter is a URL
		queueData := AudioQueueEntry{MediaURL: args[0], Requester: env.Message.Author}
		queueData.FillMetadata()
		if voiceIsStreaming(env.Guild.ID) {
			guildData[env.Guild.ID].QueueAdd(queueData)
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			return queueData.GetQueueAddedEmbed()
		}
		guildData[env.Guild.ID].AudioNowPlaying = queueData
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)
		return queueData.GetNowPlayingEmbed()
	}

	if voiceIsStreaming(env.Guild.ID) {
		if len(env.Message.Attachments) > 0 {
			for _, attachment := range env.Message.Attachments {
				queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: env.Message.Author}
				queueData.FillMetadata()
				guildData[env.Guild.ID].QueueAdd(queueData)
			}
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			return NewGenericEmbed("Voice", "Added the attached files to the guild queue.")
		}
		isPaused, _ := voiceGetPauseState(env.Guild.ID)
		if isPaused {
			_, _ = voiceResume(env.Guild.ID)
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			return NewGenericEmbed("Voice", "Resumed the audio playback.")
		}
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		return NewErrorEmbed("Voice Error", "Already playing audio.")
	}
	if len(env.Message.Attachments) > 0 {
		for _, attachment := range env.Message.Attachments {
			queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: env.Message.Author}
			queueData.FillMetadata()
			guildData[env.Guild.ID].QueueAdd(queueData)
		}
		return NewGenericEmbed("Voice", "Added the attached files to the guild queue. Use ``"+botData.CommandPrefix+"play`` to begin playback from the beginning of the queue.")
	}
	if guildData[env.Guild.ID].AudioNowPlaying.MediaURL != "" {
		queueData := guildData[env.Guild.ID].AudioNowPlaying
		queueData.FillMetadata()
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Message.ChannelID, queueData.MediaURL)
		return queueData.GetQueueAddedEmbed()
	}
	if len(guildData[env.Guild.ID].AudioQueue) > 0 {
		queueData := guildData[env.Guild.ID].AudioQueue[0]
		queueData.FillMetadata()
		guildData[env.Guild.ID].QueueRemove(0)
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Message.ChannelID, queueData.MediaURL)
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		return queueData.GetQueueAddedEmbed()
	}

	guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
	return NewGenericEmbed("Voice Error", "Some kind of strange logic flow occurred. Consider sending this to a developer.")
}

func commandStop(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			if voiceIsStreaming(env.Guild.ID) {
				voiceStop(env.Guild.ID)
				return NewGenericEmbed("Voice", "Stopped the audio playback.")
			}
			return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" to use before using the "+env.Command+" command.")
}

func commandSkip(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			if voiceIsStreaming(env.Guild.ID) {
				voiceSkip(env.Guild.ID)
				return nil
			}
			return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" to use before using the "+env.Command+" command.")
}

func commandPause(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			isPaused, err := voicePause(env.Guild.ID)
			if err != nil {
				if isPaused {
					return NewErrorEmbed("Voice Error", "Already playing audio.")
				}
				return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
			}
			return NewGenericEmbed("Voice", "Paused the audio playback.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" to use before using the "+env.Command+" command.")
}

func commandResume(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			isPaused, err := voiceResume(env.Guild.ID)
			if err != nil {
				if isPaused {
					return NewErrorEmbed("Voice Error", "Already playing audio.")
				}
				return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
			}
			return NewGenericEmbed("Voice", "Resumed the audio playback.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" to use before using the "+env.Command+" command.")
}

func commandRepeat(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch guildData[env.Guild.ID].VoiceData.RepeatLevel {
	case 0: //No repeat
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 1
		return NewGenericEmbed("Voice", "The queue will be repeated on a loop.")
	case 1: //Repeat the current queue
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 2
		return NewGenericEmbed("Voice", "The now playing entry will be repeated on a loop.")
	case 2: //Repeat what's in the now playing slot
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 0
		return NewGenericEmbed("Voice", "The queue will play through as normal.")
	}
	return nil
}

func commandShuffle(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	guildData[env.Guild.ID].VoiceData.Shuffle = !guildData[env.Guild.ID].VoiceData.Shuffle
	if guildData[env.Guild.ID].VoiceData.Shuffle {
		return NewGenericEmbed("Voice", "The queue will be shuffled around in a random order while playing.")
	}
	return NewGenericEmbed("Voice", "The queue will play through as normal.")
}

func commandYouTube(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
		//Wait for the handling of a previous playback command to finish
	}
	foundVoiceChannel := false
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			foundVoiceChannel = true
			voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
			break
		}
	}
	if !foundVoiceChannel {
		return NewErrorEmbed("YouTube Error", "You must join the voice channel to use before using the "+env.Command+" command.")
	}

	switch args[0] {
	case "search", "s":
		query := strings.Join(args[1:], " ")
		if query == "" {
			return NewErrorEmbed("YouTube Error", "You must enter a search query to use before using the "+args[0]+" command.")
		}

		if guildData[env.Guild.ID].YouTubeResults == nil {
			guildData[env.Guild.ID].YouTubeResults = make(map[string]*YouTubeResultNav)
		}

		guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID] = &YouTubeResultNav{}

		page := guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		err := page.Search(query)
		if err != nil {
			return NewErrorEmbed("YouTube Error", "There was an error getting a result for the specified query.")
		}

		commandList := botData.CommandPrefix + env.Command + " select N - Selects result N"
		if page.PrevPageToken != "" {
			commandList += botData.CommandPrefix + env.Command + "\n prev - Displays the results for the previous page"
		}
		if page.NextPageToken != "" {
			commandList += botData.CommandPrefix + env.Command + "\n next - Displays the results for the next page"
		}
		commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

		results, _ := page.GetResults()
		responseEmbed := NewEmbed().
			SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
			SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
			SetColor(0xFF0000).MessageEmbed

		fields := []*discordgo.MessageEmbedField{}
		for i := 0; i < len(results); i++ {
			videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
			if err == nil {
				author := videoInfo.Author
				title := videoInfo.Title

				fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "\"" + title + "\" by " + author})
			}
		}
		fields = append(fields, commandListField)
		responseEmbed.Fields = fields

		return responseEmbed
	case "next", "n", "forward", "+":
		if guildData[env.Guild.ID].YouTubeResults == nil {
			return NewErrorEmbed("YouTube Error", "No search session is in progress.")
		}

		page := guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		err := page.Next()
		if err != nil {
			return NewErrorEmbed("YouTube Error", "There was an error finding the next page.")
		}

		commandList := botData.CommandPrefix + env.Command + " select N - Selects result N"
		if page.PrevPageToken != "" {
			commandList += botData.CommandPrefix + env.Command + "\n prev - Displays the results for the previous page"
		}
		if page.NextPageToken != "" {
			commandList += botData.CommandPrefix + env.Command + "\n next - Displays the results for the next page"
		}
		commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

		results, _ := page.GetResults()
		responseEmbed := NewEmbed().
			SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
			SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
			SetColor(0xFF0000).MessageEmbed

		fields := []*discordgo.MessageEmbedField{}
		for i := 0; i < len(results); i++ {
			videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
			if err == nil {
				author := videoInfo.Author
				title := videoInfo.Title

				fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "\"" + title + "\" by " + author})
			}
		}
		fields = append(fields, commandListField)
		responseEmbed.Fields = fields

		return responseEmbed
	case "prev", "previous", "p", "back", "-":
		if guildData[env.Guild.ID].YouTubeResults == nil {
			return NewErrorEmbed("YouTube Error", "No search session is in progress.")
		}

		page := guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		err := page.Prev()
		if err != nil {
			return NewErrorEmbed("YouTube Error", "There was an error finding the previous page.")
		}

		commandList := botData.CommandPrefix + env.Command + " select N - Selects result N"
		if page.PrevPageToken != "" {
			commandList += botData.CommandPrefix + env.Command + "\n prev - Displays the results for the previous page"
		}
		if page.NextPageToken != "" {
			commandList += botData.CommandPrefix + env.Command + "\n next - Displays the results for the next page"
		}
		commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

		results, _ := page.GetResults()
		responseEmbed := NewEmbed().
			SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
			SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
			SetColor(0xFF0000).MessageEmbed

		fields := []*discordgo.MessageEmbedField{}
		for i := 0; i < len(results); i++ {
			videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
			if err == nil {
				author := videoInfo.Author
				title := videoInfo.Title

				fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "\"" + title + "\" by " + author})
			}
		}
		fields = append(fields, commandListField)
		responseEmbed.Fields = fields

		return responseEmbed
	case "cancel", "c":
		if guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID] != nil {
			guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID] = nil
			return NewGenericEmbed("YouTube", "Cancelled the search session.")
		}
		return NewErrorEmbed("YouTube Error", "No search session is in progress.")
	case "select", "choose":
		if guildData[env.Guild.ID].YouTubeResults == nil {
			return NewErrorEmbed("YouTube Error", "No search session is in progress.")
		}
		if len(args) < 2 {
			return NewErrorEmbed("YouTube Error", "You must specify which search result to select.")
		}

		page := guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		results, _ := page.GetResults()

		selection, err := strconv.Atoi(args[1])
		if err != nil {
			return NewErrorEmbed("YouTube Error", "``"+args[1]+"`` is not a valid number.")
		}
		if selection > len(results) || selection <= 0 {
			return NewErrorEmbed("YouTube Error", "An invalid selection was specified.")
		}

		result := results[selection-1]
		resultURL := "https://youtube.com/watch?v=" + result.Id.VideoId

		queueData := AudioQueueEntry{MediaURL: resultURL, Requester: env.Message.Author, Type: "youtube"}
		queueData.FillMetadata()
		if voiceIsStreaming(env.Guild.ID) {
			guildData[env.Guild.ID].QueueAdd(queueData)
			return queueData.GetQueueAddedEmbed()
		}
		guildData[env.Guild.ID].AudioNowPlaying = queueData
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)
		return queueData.GetNowPlayingEmbed()
	}
	return NewErrorEmbed("YouTube Error", "Unknown command ``"+args[0]+"``.")
}

func commandQueue(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(args) >= 1 {
		switch args[0] {
		case "clear":
			if len(guildData[env.Guild.ID].AudioQueue) > 0 {
				guildData[env.Guild.ID].QueueClear()
				return NewGenericEmbed("Queue", "Cleared the queue.")
			}
			return NewErrorEmbed("Queue Error", "There are no entries in the queue to clear.")
		case "remove":
			if len(args) == 1 {
				return NewErrorEmbed("Queue Error", "You must specify which queue entries to remove.")
			}

			for _, queueEntry := range args[1:] {
				queueEntryNumber, err := strconv.Atoi(queueEntry)
				if err != nil {
					return NewErrorEmbed("Queue Error", "``"+queueEntry+"`` is not a valid number.")
				}
				queueEntryNumber--

				if queueEntryNumber > len(guildData[env.Guild.ID].AudioQueue) || queueEntryNumber < 0 {
					return NewErrorEmbed("Queue Error", "``"+queueEntry+"`` is not a valid queue entry.")
				}
			}

			var newAudioQueue []AudioQueueEntry
			for queueEntryN, queueEntry := range guildData[env.Guild.ID].AudioQueue {
				keepQueueEntry := true
				for _, removedQueueEntry := range args[1:] {
					removedQueueEntryNumber, _ := strconv.Atoi(removedQueueEntry)
					removedQueueEntryNumber--
					if queueEntryN == removedQueueEntryNumber {
						keepQueueEntry = false
						break
					}
				}
				if keepQueueEntry {
					newAudioQueue = append(newAudioQueue, queueEntry)
				}
			}

			guildData[env.Guild.ID].AudioQueue = newAudioQueue

			if len(args) > 2 {
				return NewGenericEmbed("Queue", "Successfully removed the specified queue entries.")
			}
			return NewGenericEmbed("Queue", "Successfully removed the specified queue entry.")
		}
	}

	if len(guildData[env.Guild.ID].AudioQueue) == 0 {
		return NewErrorEmbed("Queue Error", "There are no entries in the queue.")
	}
	queueList := ""
	for queueEntryNumber, queueEntry := range guildData[env.Guild.ID].AudioQueue {
		displayNumber := strconv.Itoa(queueEntryNumber + 1)
		if queueList != "" {
			queueList += "\n"
		}
		switch queueEntry.Type {
		case "youtube", "soundcloud":
			queueList += displayNumber + ". [" + queueEntry.Title + "](" + queueEntry.MediaURL + ") by **" + queueEntry.Author + "** | Requested by <@" + queueEntry.Requester.ID + ">"
		default:
			queueList += displayNumber + ". " + queueEntry.MediaURL + " | Requested by <@" + queueEntry.Requester.ID + ">"
		}
	}
	return NewGenericEmbed("Queue for "+env.Guild.Name, queueList)
}

func commandNowPlaying(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if voiceIsStreaming(env.Guild.ID) {
		return guildData[env.Guild.ID].AudioNowPlaying.GetNowPlayingDurationEmbed(guildData[env.Guild.ID].VoiceData.StreamingSession)
	}
	return NewErrorEmbed("Now Playing Error", "There is no audio currently playing.")
}

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

//Debug commands
func commandBotInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	botGuilds, err := botData.DiscordSession.UserGuilds(100, "", "")
	if err != nil {
		return NewErrorEmbed("Bot Info Error", "An error occurred retrieving a list of guilds.")
	}
	botGuildNames := make([]string, 0)
	for i := 0; i < len(botGuilds); i++ {
		botGuildNames = append(botGuildNames, botGuilds[i].Name)
	}

	return NewEmbed().
		SetTitle("Bot Info").
		SetDescription("Info regarding this bot.").
		AddField("Guild List ("+strconv.Itoa(len(botGuildNames))+")", strings.Join(botGuildNames, "\n")).
		SetColor(0x1C1C1C).MessageEmbed
}
