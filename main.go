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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JoshuaDoes/duckduckgolang"
	"github.com/JoshuaDoes/go-soundcloud"
	"github.com/JoshuaDoes/go-wolfram"
	"github.com/JoshuaDoes/spotigo"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
	"github.com/koffeinsource/go-klogger"
	"github.com/mitchellh/go-ps"
	"github.com/nishanths/go-xkcd"
	"github.com/rhnvrm/lyric-api-go"
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

	//Contains guild-specific starboard data in a string map, where key = guild ID
	starboards = make(map[string]*Starboard)

	//Contains all remind entries
	remindEntries = make([]RemindEntry, 0)
)

var (
	configFile  string
	configIsBot string
	masterPID   int
	killOldBot  string
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "The location of the JSON-structured configuration file")
	flag.StringVar(&configIsBot, "bot", "false", "Whether or not to act as a bot")
	flag.IntVar(&masterPID, "masterpid", -1, "The bot master's PID")
	flag.StringVar(&killOldBot, "killold", "false", "Whether or not to kill an old bot process")
}

func main() {
	defer recoverPanic()
	debugLog("Clinet-Discord © JoshuaDoes: 2018.", true)
	debugLog("Build ID: "+BuildID, true)

	flag.Parse()
	if configIsBot == "true" {
		debugLog("Process mode: BOT", true)
	} else {
		debugLog("Process mode: MASTER", true)
	}

	debugLog("Current PID: "+strconv.Itoa(os.Getpid()), true)
	debugLog("", true)

	if configIsBot == "true" {
		numCPU := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPU * 2)
		debugLog(fmt.Sprintf("> CPU Core Count: %d / Max Process Count: %d", numCPU, numCPU*2), true)

		debugLog("> Loading settings...", true)
		configFileHandle, err := os.Open(configFile)
		defer configFileHandle.Close()
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
		if botData.BotOptions.UseSpotify {
			botData.BotClients.Spotify = &spotigo.Client{Host: botData.BotKeys.SpotifyHost, Pass: botData.BotKeys.SpotifyPass}
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
			youtubeClient, err := youtube.New(httpClient)
			if err != nil {
				debugLog("> Error initializing YouTube: "+fmt.Sprintf("%v", err), true)
			} else {
				botData.BotClients.YouTube = youtubeClient
			}
		}
		if botData.BotOptions.UseGitHub {
			botData.BotClients.GitHub = github.NewClient(nil)
		}
		if botData.BotOptions.UseLyrics {
			botData.BotClients.Lyrics = lyrics.New(lyrics.WithoutProviders(), lyrics.WithLyricsWikia(), lyrics.WithMusixMatch(), lyrics.WithSongLyrics(), lyrics.WithGeniusLyrics(botData.BotKeys.GeniusAccessToken))
		}

		debugLog("> Creating a Discord session...", true)
		discord, err := discordgo.New("Bot " + botData.BotToken)
		if err != nil {
			panic(err)
		}

		debugLog("> Registering Discord event handlers...", false)
		discord.AddHandler(discordChannelCreate)
		discord.AddHandler(discordChannelUpdate)
		discord.AddHandler(discordChannelDelete)
		discord.AddHandler(discordGuildUpdate)
		discord.AddHandler(discordGuildBanAdd)
		discord.AddHandler(discordGuildBanRemove)
		discord.AddHandler(discordGuildMemberAdd)
		discord.AddHandler(discordGuildMemberRemove)
		discord.AddHandler(discordGuildRoleCreate)
		discord.AddHandler(discordGuildRoleUpdate)
		discord.AddHandler(discordGuildRoleDelete)
		discord.AddHandler(discordGuildEmojisUpdate)
		discord.AddHandler(discordUserUpdate)
		discord.AddHandler(discordVoiceStateUpdate)
		discord.AddHandler(discordMessageCreate)
		discord.AddHandler(discordMessageDelete)
		discord.AddHandler(discordMessageDeleteBulk)
		discord.AddHandler(discordMessageUpdate)
		discord.AddHandler(discordMessageReactionAdd)
		discord.AddHandler(discordMessageReactionRemove)
		discord.AddHandler(discordMessageReactionRemoveAll)
		discord.AddHandler(discordReady)

		//If a state exists, restore it
		debugLog("> Restoring state...", false)
		stateRestore()

		debugLog("> Connecting to Discord...", true)
		err = discord.Open()
		if err != nil {
			panic(err)
		}
		debugLog("> Connection successful", true)
		botData.DiscordSession = discord

		if botData.SendOwnerStackTraces {
			checkPanicRecovery()
		}

		debugLog("> Checking if bot was restarted...", false)
		checkRestart()

		debugLog("> Checking if bot was updated...", false)
		checkUpdate()

		debugLog("> Halting main() until SIGINT, SIGTERM, INTERRUPT, or KILL", false)
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill, syscall.SIGKILL)
		<-sc

		//Save the current state before shutting down
		// Note: This is done before shutting down as the shutdown process may yield
		//       some errors with goroutines like voice playback
		stateSave()

		for guildID, guildDataRow := range guildData {
			if guildDataRow.VoiceData.VoiceConnection != nil {
				if voiceIsStreaming(guildID) {
					if botData.Updating {
						//Notify users that an update is occuring
						botData.DiscordSession.ChannelMessageSendEmbed(guildDataRow.VoiceData.ChannelIDJoinedFrom, NewEmbed().SetTitle("Update").SetDescription("Your audio playback has been interrupted for a "+botData.BotName+" update event. You may resume playback in a few seconds.").SetColor(0x1C1C1C).MessageEmbed)
					}
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
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		watchdogTicker := time.Tick(1 * time.Second)

		for {
			select {
			case _, ok := <-sc:
				if ok {
					botProcess, _ := os.FindProcess(botPid)
					_ = botProcess.Signal(syscall.SIGKILL)
					os.Exit(0)
				}
			case <-watchdogTicker:
				if !isProcessRunning(botPid) {
					botPid = spawnBot()
				}
			}
		}
	}
}

func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	defer recoverPanic()

	debugLog("> Setting bot username from Discord...", false)
	botData.BotName = session.State.User.Username

	debugLog("> Preparing command list...", false)
	initCommands()

	debugLog("> Preparing voice services...", false)
	initVoiceServices()

	debugLog("> Setting random status...", false)
	updateRandomStatus(session, 0)

	debugLog("> Creating cronjob session...", false)
	cronjob := cron.New()

	debugLog("> Creating random status update cronjob...", false)
	cronjob.AddFunc("@every 1m", func() { updateRandomStatus(session, 0) })

	debugLog("> Creating tip message cronjob...", false)
	cronjob.AddFunc("@every 1h", func() { sendTipMessages() })

	debugLog("> Starting cronjobs...", false)
	cronjob.Start()

	debugLog("> Preparing saved remind entries...", false)
	oldRemindEntries := remindEntries
	remindEntries = make([]RemindEntry, 0)
	for i := range oldRemindEntries {
		remindWhen(oldRemindEntries[i].UserID, oldRemindEntries[i].GuildID, oldRemindEntries[i].ChannelID, oldRemindEntries[i].Message, oldRemindEntries[i].Added, oldRemindEntries[i].When, time.Now())
	}
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

func sendTipMessages() {
	tipMessageN := -1
	for {
		tipMessageN = rand.Intn(len(botData.TipMessages))
		if tipMessageN != botData.LastTipMessage {
			break
		}
	}
	tipMessage := botData.TipMessages[tipMessageN]
	tipMessageEmbed := NewEmbed().
		AddField("Did You Know?", tipMessage.DidYouKnow).
		AddField("How To Use", tipMessage.HowTo).
		SetFooter("Feature: " + tipMessage.FeatureName).
		SetColor(0x1C1C1C)

	if len(tipMessage.Examples) > 0 {
		tipMessageEmbed.AddField("Examples", strings.Join(tipMessage.Examples, "\n"))
	}

	for _, guild := range guildSettings {
		if guild.TipsChannel != "" {
			botData.DiscordSession.ChannelMessageSendEmbed(guild.TipsChannel, tipMessageEmbed.MessageEmbed)
		}
	}

	botData.LastTipMessage = tipMessageN
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
	guildDataJSON, err := json.MarshalIndent(guildData, "", "\t")
	if err != nil {
		debugLog("> Error saving guildData state: "+err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/guildData.json", guildDataJSON, 0644)
		if err != nil {
			debugLog("> Error saving guildData state: "+err.Error(), true)
		}
	}

	guildSettingsJSON, err := json.MarshalIndent(guildSettings, "", "\t")
	if err != nil {
		debugLog("> Error saving guildSettings state: "+err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/guildSettings.json", guildSettingsJSON, 0644)
		if err != nil {
			debugLog("> Error saving guildSettings state: "+err.Error(), true)
		}
	}

	userSettingsJSON, err := json.MarshalIndent(userSettings, "", "\t")
	if err != nil {
		debugLog("> Error saving userSettings state: "+err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/userSettings.json", userSettingsJSON, 0644)
		if err != nil {
			debugLog("> Error saving userSettings state: "+err.Error(), true)
		}
	}

	starboardsJSON, err := json.MarshalIndent(starboards, "", "\t")
	if err != nil {
		debugLog("> Error saving starboards state: "+err.Error(), true)
		debugLog(err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/starboards.json", starboardsJSON, 0644)
		if err != nil {
			debugLog("> Error saving starboards state: "+err.Error(), true)
		}
	}

	remindEntriesJSON, err := json.MarshalIndent(remindEntries, "", "\t")
	if err != nil {
		debugLog("> Error saving remind entries: "+err.Error(), true)
	} else {
		err = ioutil.WriteFile("state/reminds.json", remindEntriesJSON, 0644)
		if err != nil {
			debugLog("> Error saving remind entries: "+err.Error(), true)
		}
	}
}

func stateRestore() {
	guildDataJSON, err := ioutil.ReadFile("state/guildData.json")
	if err == nil {
		err = json.Unmarshal(guildDataJSON, &guildData)
		if err != nil {
			debugLog("> Error restoring guildData state: "+err.Error(), true)
		}
	} else {
		debugLog("> No guildData state was found", true)
	}

	guildSettingsJSON, err := ioutil.ReadFile("state/guildSettings.json")
	if err == nil {
		err = json.Unmarshal(guildSettingsJSON, &guildSettings)
		if err != nil {
			debugLog("> Error restoring guildSettings state: "+err.Error(), true)
		}
	} else {
		debugLog("> No guildSettings state was found", true)
	}

	userSettingsJSON, err := ioutil.ReadFile("state/userSettings.json")
	if err == nil {
		err = json.Unmarshal(userSettingsJSON, &userSettings)
		if err != nil {
			debugLog("> Error restoring userSettings state: "+err.Error(), true)
		}
	} else {
		debugLog("> No userSettings state was found", true)
	}

	starboardsJSON, err := ioutil.ReadFile("state/starboards.json")
	if err == nil {
		err = json.Unmarshal(starboardsJSON, &starboards)
		if err != nil {
			debugLog("> Error restoring starboards state: "+err.Error(), true)
		}
	} else {
		debugLog("> No starboards state was found", true)
	}

	remindEntriesJSON, err := ioutil.ReadFile("state/reminds.json")
	if err == nil {
		err = json.Unmarshal(remindEntriesJSON, &remindEntries)
		if err != nil {
			debugLog("> Error restoring remind entries: "+err.Error(), true)
		}
	} else {
		debugLog("> No remind entries were found", true)
	}
}

func recoverPanic() {
	if panicReason := recover(); panicReason != nil {
		fmt.Println("Clinet has encountered an unrecoverable error and has crashed.")
		fmt.Println("Some information describing this crash: " + panicReason.(error).Error())
		if botData.SendOwnerStackTraces || configIsBot == "false" {
			stack := make([]byte, 65536)
			l := runtime.Stack(stack, true)
			fmt.Println("Stack trace:\n" + string(stack[:l]))
			err := ioutil.WriteFile("stacktrace.txt", stack[:l], 0644)
			if err != nil {
				fmt.Println("Failed to write stack trace.")
			}
			err = ioutil.WriteFile("crash.txt", []byte(panicReason.(error).Error()), 0644)
			if err != nil {
				fmt.Println("Failed to write crash error.")
			}
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
			botData.DiscordSession.ChannelFileSendWithMessage(ownerPrivChannelID, "Stack trace:", "stacktrace.txt", stack)
		}

		stack.Close()
		os.Remove("crash.txt")
		os.Remove("stacktrace.txt")
	}
}

func checkRestart() {
	restartChannelID, err := ioutil.ReadFile(".restart")
	if err == nil && len(restartChannelID) > 0 {
		restartEmbed := NewGenericEmbed("Restart", "Successfully restarted "+botData.BotName+"!")
		botData.DiscordSession.ChannelMessageSendEmbed(string(restartChannelID), restartEmbed)

		os.Remove(".restart")
	}
}

func checkUpdate() {
	updateChannelID, err := ioutil.ReadFile(".update")
	if err == nil && len(updateChannelID) > 0 {
		updateEmbed := NewGenericEmbed("Update", "Successfully updated "+botData.BotName+"!")
		botData.DiscordSession.ChannelMessageSendEmbed(string(updateChannelID), updateEmbed)

		os.Remove(".update")
	}
}

func spawnBot() int {
	if killOldBot == "true" {
		processList, err := ps.Processes()
		if err == nil {
			for _, process := range processList {
				if process.Pid() != os.Getpid() && process.Pid() != masterPID && process.Executable() == filepath.Base(os.Args[0]) {
					oldProcess, err := os.FindProcess(process.Pid())
					if err == nil {
						oldProcess.Signal(syscall.SIGKILL)
					}
				}
			}
		}
	}
	os.Remove(os.Args[0] + ".old")

	botProcess := exec.Command(os.Args[0], "-bot", "true", "-masterpid", strconv.Itoa(os.Getpid()))
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
