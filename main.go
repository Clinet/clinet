package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/JoshuaDoes/duckduckgolang"
	"github.com/JoshuaDoes/go-soundcloud"
	"github.com/JoshuaDoes/go-wolfram"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
	"github.com/koffeinsource/go-klogger"
	"github.com/nishanths/go-xkcd"
	"github.com/robfig/cron"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

var (
	//Contains all bot configurations
	botData *BotData = &BotData{}

	//Contains guild-specific data in a string map, where key = guild ID
	guildData = make(map[string]*GuildData)

	//Contains guild-specific settings in a string map, where key = guild ID
	guildSettings = make(map[string]*GuildSettings)

	//Contains user-specific settings in a string map, where key = user ID
	userSettings = make(map[string]*UserSettings)
)

var (
	configFile  string
	configIsBot string
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "The location of the JSON-structured configuration file")
	flag.StringVar(&configIsBot, "bot", "false", "Whether or not to act as a bot")
}

func main() {
	defer recoverPanic()
	debugLog("Clinet-Discord © JoshuaDoes: 2018.", true)
	debugLog("Build ID: "+BuildID+"\n", true)

	flag.Parse()
	if configIsBot == "true" {
		debugLog("Process mode: BOT", true)
	} else {
		debugLog("Process mode: MASTER", true)
	}

	if configIsBot == "true" {
		debugLog("> Loading settings...", true)
		configFileHandle, err := os.Open(configFile)
		if err != nil {
			panic("Error loading configuration file `" + configFile + "`")
		} else {
			configParser := json.NewDecoder(configFileHandle)
			if err = configParser.Decode(&botData); err != nil {
				panic(err)
			} else {
				configErr := botData.PrepConfig() //Check the configuration for any errors or inconsistencies, then prepare it for usage
				if configErr != nil {
					panic(configErr)
				}
			}
		}

		debugLog("> Initializing clients for external services...", true)
		if botData.BotOptions.UseDuckDuckGo {
			debugLog("> Initializing DuckDuckGo...", false)
			botData.BotClients.DuckDuckGo = &duckduckgo.Client{AppName: botData.BotKeys.DuckDuckGoAppName}
		}
		if botData.BotOptions.UseImgur {
			debugLog("> Initializing Imgur HTTP client...", false)
			botData.BotClients.Imgur.HTTPClient = &http.Client{}
			debugLog("> Initializing Imgur CLILogger...", false)
			botData.BotClients.Imgur.Log = &klogger.CLILogger{}
			debugLog("> Initializing Imgur...", false)
			botData.BotClients.Imgur.ImgurClientID = botData.BotKeys.ImgurClientID
		}
		if botData.BotOptions.UseSoundCloud {
			debugLog("> Initializing SoundCloud...", false)
			botData.BotClients.SoundCloud = &soundcloud.Client{ClientID: botData.BotKeys.SoundCloudClientID, AppVersion: botData.BotKeys.SoundCloudAppVersion}
		}
		if botData.BotOptions.UseWolframAlpha {
			debugLog("> Initializing Wolfram|Alpha...", false)
			botData.BotClients.Wolfram = &wolfram.Client{AppID: botData.BotKeys.WolframAppID}
		}
		if botData.BotOptions.UseXKCD {
			debugLog("> Initializing XKCD...", false)
			botData.BotClients.XKCD = xkcd.NewClient()
		}
		if botData.BotOptions.UseYouTube {
			debugLog("> Initializing YouTube...", false)
			httpClient := &http.Client{
				Transport: &transport.APIKey{Key: botData.BotKeys.YouTubeAPIKey},
			}
			youtubeClient, err := youtube.New(httpClient)
			if err != nil {
				debugLog("> Error initializing YouTube", true)
				debugLog("Error: "+fmt.Sprintf("%v", err), false)
			} else {
				botData.BotClients.YouTube = youtubeClient
			}
		}
		if botData.BotOptions.UseGitHub {
			debugLog("> Initializing GitHub...", false)
			botData.BotClients.GitHub = github.NewClient(nil)
		}

		debugLog("> Preparing command list...", true)
		initCommands()

		debugLog("> Creating a Discord session...", true)
		discord, err := discordgo.New("Bot " + botData.BotToken)
		if err != nil {
			panic(err)
		}

		debugLog("> Registering Discord event handlers...", false)
		discord.AddHandler(discordMessageCreate)
		discord.AddHandler(discordMessageDelete)
		discord.AddHandler(discordMessageDeleteBulk)
		discord.AddHandler(discordMessageUpdate)
		discord.AddHandler(discordUserJoin)
		discord.AddHandler(discordUserLeave)
		discord.AddHandler(discordReady)

		//If a state exists, restore it
		stateRestore()

		debugLog("> Connecting to Discord...", true)
		err = discord.Open()
		if err != nil {
			panic(err)
		}
		debugLog("> Connection successful", true)
		botData.DiscordSession = discord

		checkPanicRecovery()

		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc

		//Save the current state before shutting down
		// Note: This is done before shutting down as the shutdown process may yield
		//       some errors with goroutines like voice playback
		stateSave()

		for guildID, guildDataRow := range guildData {
			if guildDataRow.VoiceData.VoiceConnection != nil {
				if voiceIsStreaming(guildID) {
					debugLog("> Stopping stream in voice channel "+guildDataRow.VoiceData.VoiceConnection.ChannelID+"...", false)
					voiceStop(guildID)
				}
				debugLog("> Closing connection to voice channel "+guildDataRow.VoiceData.VoiceConnection.ChannelID+"...", false)
				guildDataRow.VoiceData.VoiceConnection.Close()
			}
		}
		debugLog("> Disconnecting from Discord...", true)
		discord.Close()
	} else {
		botPid := spawnBot()
		sc := make(chan os.Signal, 1)

		for {
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
			select {
			case _, ok := <-sc:
				if ok {
					botProcess, _ := os.FindProcess(botPid)
					_ = botProcess.Kill()
					os.Exit(0)
				}
			default:
				if !isProcessRunning(botPid) {
					botPid = spawnBot()
				}
			}
		}
	}
}

func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	defer recoverPanic()

	updateRandomStatus(session, 0)
	cronjob := cron.New()
	cronjob.AddFunc("@every 1m", func() { updateRandomStatus(session, 0) })
	cronjob.Start()
}

func updateRandomStatus(session *discordgo.Session, status int) {
	if status == 0 {
		status = rand.Intn(len(botData.CustomStatuses)) + 1
	}
	status--

	switch botData.CustomStatuses[status].Type {
	case 0:
		session.UpdateStatus(0, botData.CustomStatuses[status].Status)
	case 1:
		session.UpdateListeningStatus(botData.CustomStatuses[status].Status)
	case 2:
		session.UpdateStreamingStatus(0, botData.CustomStatuses[status].Status, botData.CustomStatuses[status].URL)
	}
}

func typingEvent(session *discordgo.Session, channelID string) {
	if botData.BotOptions.SendTypingEvent {
		session.ChannelTyping(channelID)
	}
}

func debugLog(msg string, overrideConfig bool) {
	if botData.DebugMode || overrideConfig {
		fmt.Println(msg)
	}
}

func stateSave() {
	debugLog("> Saving guildData state...", true)
	guildDataJSON, err := json.MarshalIndent(guildData, "", "\t")
	if err != nil {
		debugLog("> Error saving guildData state", true)
		debugLog(err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/guildData.json", guildDataJSON, 0644)
		if err != nil {
			debugLog("> Error saving guildData state", true)
			debugLog(err.Error(), true)
		}
	}

	debugLog("> Saving guildSettings state...", true)
	guildSettingsJSON, err := json.MarshalIndent(guildSettings, "", "\t")
	if err != nil {
		debugLog("> Error saving guildSettings state", true)
		debugLog(err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/guildSettings.json", guildSettingsJSON, 0644)
		if err != nil {
			debugLog("> Error saving guildSettings state", true)
			debugLog(err.Error(), true)
		}
	}

	debugLog("> Saving userSettings state...", true)
	userSettingsJSON, err := json.MarshalIndent(userSettings, "", "\t")
	if err != nil {
		debugLog("> Error saving userSettings state", true)
		debugLog(err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/userSettings.json", userSettingsJSON, 0644)
		if err != nil {
			debugLog("> Error saving userSettings state", true)
			debugLog(err.Error(), true)
		}
	}
}

func stateRestore() {
	guildDataJSON, err := ioutil.ReadFile("state/guildData.json")
	if err == nil {
		debugLog("> Restoring guildData state...", true)
		err = json.Unmarshal(guildDataJSON, &guildData)
		if err != nil {
			debugLog("> Error restoring state", true)
			debugLog(err.Error(), true)
		}
	} else {
		debugLog("> No guildData state was found", true)
	}

	guildSettingsJSON, err := ioutil.ReadFile("state/guildSettings.json")
	if err == nil {
		debugLog("> Restoring guildSettings state...", true)
		err = json.Unmarshal(guildSettingsJSON, &guildSettings)
		if err != nil {
			debugLog("> Error restoring state", true)
			debugLog(err.Error(), true)
		}
	} else {
		debugLog("> No guildSettings state was found", true)
	}

	userSettingsJSON, err := ioutil.ReadFile("state/userSettings.json")
	if err == nil {
		debugLog("> Restoring userSettings state...", true)
		err = json.Unmarshal(userSettingsJSON, &userSettings)
		if err != nil {
			debugLog("> Error restoring state", true)
			debugLog(err.Error(), true)
		}
	} else {
		debugLog("> No userSettings state was found", true)
	}
}

func recoverPanic() {
	if panicReason := recover(); panicReason != nil {
		fmt.Println("Clinet has encountered an unrecoverable error and has crashed.")
		fmt.Println("Some information describing this crash: " + panicReason.(error).Error())
		stack := make([]byte, 65536)
		l := runtime.Stack(stack, true)
		err := ioutil.WriteFile("stacktrace.txt", stack[:l], 0644)
		if err != nil {
			fmt.Println("Failed to write stack trace.")
		}
		err = ioutil.WriteFile("crash.txt", []byte(panicReason.(error).Error()), 0644)
		if err != nil {
			fmt.Println("Failed to write crash error.")
		}
		os.Exit(1)
	}
}

func checkPanicRecovery() {
	ownerPrivChannel, err := botData.DiscordSession.UserChannelCreate(botData.BotOwnerID)
	if err != nil {
		debugLog("An error occurred creating a private channel with the bot owner.", false)
	} else {
		ownerPrivChannelID := ownerPrivChannel.ID

		crash, crashErr := ioutil.ReadFile("crash.txt")
		stack, stackErr := os.Open("stacktrace.txt")

		if crashErr == nil && stackErr == nil {
			botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, "Clinet has just recovered from an error that caused a crash.")
			botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, "Crash:\n```"+string(crash)+"```")
			//botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, string(stack))
			botData.DiscordSession.ChannelFileSendWithMessage(ownerPrivChannelID, "Stack trace:", "stacktrace.txt", stack)
		}

		stack.Close()
		os.Remove("crash.txt")
		os.Remove("stacktrace.txt")
	}
}

func spawnBot() int {
	botProcess := exec.Command(os.Args[0], "-bot", "true")
	botProcess.Stdout = os.Stdout
	botProcess.Stderr = os.Stderr
	err := botProcess.Start()
	if err != nil {
		panic(err)
	}
	return botProcess.Process.Pid
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if runtime.GOOS != "windows" {
		return process.Signal(syscall.Signal(0)) == nil
	}

	processState, err := process.Wait()
	if err != nil {
		return false
	}
	if processState.Exited() {
		return false
	}

	return true
}
