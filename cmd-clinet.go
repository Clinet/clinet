package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JoshuaDoes/duckduckgolang"
	"github.com/JoshuaDoes/go-soundcloud"
	"github.com/JoshuaDoes/go-wolfram"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
	"github.com/koffeinsource/go-klogger"
	"github.com/nishanths/go-xkcd"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"gopkg.in/src-d/go-git.v4"
)

func commandBotInfo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	guildCount := len(botData.DiscordSession.State.Guilds)
	commandCount := len(botData.Commands)

	return NewEmbed().
		SetTitle("Bot Info").
		SetDescription("Info regarding this bot.").
		AddField("Bot Name", botData.BotName).
		AddField("Bot Owner", "<@!"+botData.BotOwnerID+">").
		AddField("Guild Count", strconv.Itoa(guildCount)).
		AddField("Default Prefix", botData.CommandPrefix).
		AddField("Command Count", strconv.Itoa(commandCount)).
		AddField("Disabled Wolfram|Alpha Pods", strings.Join(botData.BotOptions.WolframDeniedPods, ", ")).
		AddField("Debug Mode", strconv.FormatBool(botData.DebugMode)).
		InlineAllFields().
		SetColor(0x1C1C1C).MessageEmbed
}

func commandReload(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	configFileHandle, err := os.Open(configFile)
	defer configFileHandle.Close()
	if err != nil {
		return NewErrorEmbed("Reload Error", "There was an error loading the bot configuration file.")
	}

	configParser := json.NewDecoder(configFileHandle)
	if err = configParser.Decode(&botData); err != nil {
		return NewErrorEmbed("Reload Error", "There was an error applying the bot configuration to memory. State and/or configuration may be corrupted, consider checking the configuration and restarting the bot process.")
	}

	err = botData.PrepConfig()
	if err != nil {
		return NewErrorEmbed("Reload Error", "There were some inconsistencies with the bot configuration. State and/or configuration may be corrupted, consider checking the configuration and restarting the bot process.")
	}

	if botData.BotOptions.UseDuckDuckGo {
		botData.BotClients.DuckDuckGo = &duckduckgo.Client{AppName: botData.BotKeys.DuckDuckGoAppName}
	}
	if botData.BotOptions.UseImgur {
		botData.BotClients.Imgur.HTTPClient = &http.Client{}
		botData.BotClients.Imgur.Log = &klogger.CLILogger{}
		botData.BotClients.Imgur.ImgurClientID = botData.BotKeys.ImgurClientID
	}
	if botData.BotOptions.UseSoundCloud {
		botData.BotClients.SoundCloud = &soundcloud.Client{ClientID: botData.BotKeys.SoundCloudClientID, AppVersion: botData.BotKeys.SoundCloudAppVersion}
	}
	if botData.BotOptions.UseWolframAlpha {
		botData.BotClients.Wolfram = &wolfram.Client{AppID: botData.BotKeys.WolframAppID}
	}
	if botData.BotOptions.UseXKCD {
		botData.BotClients.XKCD = xkcd.NewClient()
	}
	if botData.BotOptions.UseYouTube {
		httpClient := &http.Client{
			Transport: &transport.APIKey{Key: botData.BotKeys.YouTubeAPIKey},
		}
		botData.BotClients.YouTube, _ = youtube.New(httpClient)
	}
	if botData.BotOptions.UseGitHub {
		botData.BotClients.GitHub = github.NewClient(nil)
	}

	return NewGenericEmbed("Reload", "Successfully reloaded the bot configuration.")
}

func commandRestart(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//Tell the user we're restarting
	botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, NewGenericEmbed("Restart", "Restarting "+botData.BotName+"..."))

	//Write the current channel ID to a restart file for the bot to read after the restart
	ioutil.WriteFile(".restart", []byte(env.Channel.ID), 0644)

	//Save the state so it's not lost
	stateSave()

	//Close the bot process, as the MASTER process will open it again
	os.Exit(0)

	return nil
}

func commandUpdate(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//Check if the Go compiler is installed
	golangver := exec.Command("go", "version")

	output, err := golangver.CombinedOutput()
	if len(output) <= 0 || err != nil {
		return NewErrorEmbed("Update Error", "Unable to execute ``go version``. Make sure Go ["+GolangVersion+"] is installed on the host machine.\n\n"+fmt.Sprintf("%s\n```%v```", output, err))
	}

	//Check if the govvv wrapper is installed
	govvv := exec.Command("govvv")

	output, _ = govvv.CombinedOutput()
	if len(output) <= 0 {
		return NewErrorEmbed("Update Error", "Unable to execute ``govvv``. Make sure govvv is installed on the host machine."+fmt.Sprintf("```%v```", output))
	}

	//Create a temporary directory to store the git repository in
	repodir, err := ioutil.TempDir("", "clinet")
	if err != nil {
		return NewErrorEmbed("Update Error", "Error creating a temporary directory to store the Clinet git repository in.")
	}
	defer os.RemoveAll(repodir)

	//Check if an update is available
	repo, err := git.PlainClone(repodir, false, &git.CloneOptions{
		URL:   "https://github.com/JoshuaDoes/clinet",
		Depth: 1,
	})
	if err != nil {
		return NewErrorEmbed("Update Error", "Error cloning the git repo.")
	}
	ref, err := repo.Head()
	if err != nil {
		return NewErrorEmbed("Update Error", "Error finding the HEAD of the git repo.")
	}
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return NewErrorEmbed("Update Error", "Error fetching the HEAD commit of the git repo.")
	}
	commitHash := commit.Hash.String()
	if commitHash == GitCommitFull {
		if len(args) <= 0 || len(args) >= 1 && args[0] != "force" {
			return NewGenericEmbed("Update", botData.BotName+" is already up to date!")
		}
	}

	//Tell the user we're updating
	botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, NewGenericEmbed("Update", "Updating "+botData.BotName+" to commit ``"+commitHash+"`` from commit ``"+GitCommitFull+"``..."))

	//Build the update
	outputFile := filepath.Dir(repodir)
	if runtime.GOOS == "windows" {
		outputFile += ".exe"
	}

	govvvbuild := exec.Command("govvv", "build", "-o", outputFile)
	govvvbuild.Dir = repodir
	if !botData.DebugMode {
		govvvbuild.Args = append(govvvbuild.Args, "-ldflags=-s -w")
	}

	output, err = govvvbuild.CombinedOutput()
	if err != nil {
		return NewErrorEmbed("Update Error", "Unable to build "+botData.BotName+" ``"+commitHash+"``.\n\n"+fmt.Sprintf("```%s```", output))
	}

	if _, err = os.Stat(outputFile); os.IsNotExist(err) {
		return NewErrorEmbed("Update Error", "Unable to find the updated build of "+botData.BotName+" ``"+commitHash+"``.\n\n"+fmt.Sprintf("```%v```", err))
	}

	os.Rename(os.Args[0], os.Args[0]+".old")
	os.Rename(outputFile, os.Args[0])

	//Write the current channel ID to an update file for the bot to read after restarting
	ioutil.WriteFile(".update", []byte(env.Channel.ID), 0644)

	//Write the current master PID to an oldmpid file for the bot to read after restarting
	ioutil.WriteFile(".oldmpid", []byte(strconv.Itoa(masterPID)), 0644)

	//Write the current bot PID to an oldbpid file for the bot to read after restarting
	ioutil.WriteFile(".oldbpid", []byte(strconv.Itoa(os.Getpid())), 0644)

	//Save the state so it's not lost
	stateSave()

	//Spawn a new bot process that will kill this one
	botProcess := exec.Command(os.Args[0], "-bot", "true", "-killold", "true")
	botProcess.Stdout = os.Stdout
	botProcess.Stderr = os.Stderr
	err = botProcess.Start()
	if err != nil {
		return NewErrorEmbed("Update Error", "Unable to spawn the updated bot process.")
	}

	return NewGenericEmbed("Update", "Waiting for update to finish...")
}

func commandAbout(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - About").
		SetDescription(botData.BotName+" is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
		AddField("How can I use "+botData.BotName+" in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
		AddField("How can I help keep "+botData.BotName+" running?", "The best ways to help keep "+botData.BotName+" running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
		AddField("How can I use "+botData.BotName+"?", "There are many ways to make use of "+botData.BotName+".\n1) Type ``"+botData.CommandPrefix+"help`` and try using some of the available commands.\n2) Ask "+botData.BotName+" a question, ex: ``@"+botData.DiscordSession.State.User.String()+", what time is it?`` or ``@"+botData.DiscordSession.State.User.String()+", what is DiscordApp?``.").
		AddField("Where can I join the "+botData.BotName+" Discord server?", "If you would like to get help and support with "+botData.BotName+" or experiment with the latest and greatest of "+botData.BotName+", use the Discord Server Invite Link at the end of this message.").
		AddField("Bot Invite Link", botData.BotInviteURL).
		AddField("Discord Server Invite Link", botData.BotDiscordURL).
		AddField("Donation Link", botData.BotDonationURL).
		AddField("Source Code Link", botData.BotSourceURL).
		SetColor(0x1C1C1C).MessageEmbed
}
func commandInvite(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Invite").
		SetDescription("Below are the available invite links for "+botData.BotName+".").
		AddField("Bot Invite", botData.BotInviteURL).
		AddField("Discord Server (Support/Development/Testing)", botData.BotDiscordURL).
		SetColor(0x1C1C1C).MessageEmbed
}
func commandDonate(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Donate").
		SetDescription("Below are the available donation links for "+botData.BotName+".").
		AddField("PayPal", botData.BotDonationURL).
		SetColor(0x1C1C1C).MessageEmbed
}
func commandSource(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Source Code").
		SetDescription("Below are the available source code links for "+botData.BotName+".").
		AddField("GitHub", botData.BotSourceURL).
		SetColor(0x1C1C1C).MessageEmbed
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
			if permissionsAllowed, _ := MemberHasPermission(botData.DiscordSession, env.Guild.ID, env.User.ID, env.Channel.ID, discordgo.PermissionAdministrator|command.RequiredPermissions); permissionsAllowed || command.RequiredPermissions == 0 {
				commandField := &discordgo.MessageEmbedField{Name: botData.CommandPrefix + commandName, Value: command.HelpText, Inline: true}
				commandFields = append(commandFields, commandField)
			}
		}
	}

	pageNumber := 1
	if len(args) > 0 {
		newPageNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return NewErrorEmbed("Help Error", "Invalid command or page number.")
		}
		pageNumber = newPageNumber
	}

	//Create the help page and give it the command list
	helpEmbed, totalPages, err := page(commandFields, pageNumber, botData.BotOptions.HelpMaxResults)
	if err != nil {
		return NewErrorEmbed("Help Error", fmt.Sprintf("%v", err))
	}

	//Prepare the help page to look nice
	helpEmbed.
		SetTitle(botData.BotName + " - Help").
		SetDescription("A list of commands you have permission to use.").
		SetFooter("Page " + strconv.Itoa(pageNumber) + " of " + strconv.Itoa(totalPages) + " | " + botData.CommandPrefix + env.Command + " {page}").
		SetColor(0xFAFAFA)

	//Return the help page to the caller
	return helpEmbed.MessageEmbed
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
		AddField("Golang Libraries", "[duckduckgolang](https://github.com/JoshuaDoes/duckduckgolang), "+
			"[go-soundcloud](https://github.com/JoshuaDoes/go-soundcloud), "+
			"[go-wolfram](https://github.com/JoshuaDoes/go-wolfram), "+
			"[discordgo](https://github.com/bwmarrin/discordgo), "+
			"[github](https://github.com/google/go-github/github), "+
			"[dca](https://github.com/jonas747/dca), "+
			"[go-imgur](https://github.com/koffeinsource/go-imgur), "+
			"[go-klogger](https://github.com/koffeinsource/go-klogger), "+
			"[go-xkcd](https://github.com/nishanths/go-xkcd), "+
			"[cron](https://github.com/robfig/cron), "+
			"[ytdl](https://github.com/rylio/ytdl), "+
			"[transport](https://google.golang.org/api/googleapi/transport), "+
			"[youtube](https://google.golang.org/api/youtube/v3), "+
			"[urbandictionary](https://github.com/JoshuaDoes/urbandictionary), "+
			"[goprobe](https://github.com/JoshuaDoes/goprobe), "+
			"[go-cve](https://github.com/JoshuaDoes/go-cve), "+
			"[goeip](https://github.com/JoshuaDoes/goeip), "+
			"[prose](https://gopkg.in/jdkato/prose.v2), "+
			"[structs](https://github.com/fatih/structs)").
		AddField("Icon Design", "- thejsa").
		AddField("Source Code", "- https://github.com/JoshuaDoes/clinet").
		SetColor(0x1C1C1C).MessageEmbed
}

func commandPing(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//Create a list of ping test results
	pingResults := make([]int, botData.BotOptions.MaxPingCount)
	pingResultsStr := make([]string, botData.BotOptions.MaxPingCount)

	//Create ping embed
	pingEmbed := NewGenericEmbed("Ping!", "Waiting for ping...")

	//Loop through each slice entry of pingResults to store our results
	for i := 0; i < len(pingResults); i++ {
		//Get current time in milliseconds
		timeCurrent := int(time.Now().UnixNano() / 1000000)

		//Send ping embed
		pingMessage, err := botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, pingEmbed)
		if err != nil {
			pingResults[i] = -1
			continue
		}

		//Get new current time in milliseconds
		timeNew := int(time.Now().UnixNano() / 1000000)

		//Delete pingMessage to prevent spam
		botData.DiscordSession.ChannelMessageDelete(env.Channel.ID, pingMessage.ID)

		//Subtract new time from old time to get the ping
		timeDiff := timeNew - timeCurrent

		//Store the time difference in the ping results
		pingResults[i] = timeDiff
		pingResultsStr[i] = fmt.Sprintf("%dms", timeDiff)
	}

	//Determine the average ping
	pingSum := 0
	jitterCount := 0
	for i := 0; i < len(pingResults); i++ {
		if pingResults[i] == -1 {
			jitterCount++
		} else {
			pingSum += pingResults[i]
		}
	}
	pingAverage := int(pingSum / len(pingResults))

	//Create an addon message
	var addonMessage string

	//Give the addon message a random message based on the ping
	if pingAverage < 10 {
		addonMessage = "~~This bot might be running on steroids.~~"
	} else if pingAverage < 50 {
		addonMessage = "Someone take my coffee away!"
	} else if pingAverage < 100 {
		addonMessage = "Give me my coffee back. :c"
	} else if pingAverage < 150 {
		addonMessage = "Need... coffee..."
	} else if pingAverage < 200 {
		addonMessage = "I could really use some help here."
	} else {
		addonMessage = "Why am I walking again?"
	}

	//Return ping results
	return NewEmbed().
		SetTitle("Pong!").
		SetDescription(fmt.Sprintf("Average ping is ``%dms``. A total of ``%d/%d`` ping tests failed.\n*%s*", pingAverage, jitterCount, botData.BotOptions.MaxPingCount, addonMessage)).
		AddField("Ping Results", strings.Join(pingResultsStr, ", ")).
		SetFooter(fmt.Sprintf("Ping results are determined by sending %d messages and determining how long it takes for each message to send successfully and return a success code. The average ping is determined by taking the sum of all of the ping results and dividing it by %d.", botData.BotOptions.MaxPingCount, botData.BotOptions.MaxPingCount)).
		SetColor(0x1C1C1C).MessageEmbed
}
