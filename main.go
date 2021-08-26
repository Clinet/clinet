package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mmcdole/gofeed"

	duckduckgo "github.com/JoshuaDoes/duckduckgolang"
	soundcloud "github.com/JoshuaDoes/go-soundcloud"
	wolfram "github.com/JoshuaDoes/go-wolfram"
	gassist "github.com/JoshuaDoes/google-assistant/v1alpha2"
	"github.com/JoshuaDoes/spotigo"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
	klogger "github.com/koffeinsource/go-klogger"
	xkcd "github.com/nishanths/go-xkcd"
	lyrics "github.com/rhnvrm/lyric-api-go"
	"github.com/robfig/cron"
	ytdl "github.com/kkdai/youtube/v2"
	"github.com/superwhiskers/fennel"
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

	//Contains guild-specific voice data in a string map, where key = guild ID
	voiceData = make(map[string]*Voice)

	//Contains a pointer to the current log file
	logFile *os.File

	//Contains the current uptime
	uptime time.Time

	//Whether or not discordReady() has been called
	isReady bool
)

var (
	configFile       string
	gcpAuthTokenFile string
	gcpToken         *gassist.Token

	configIsBot string
	masterPID   int
	killOldBot  string
	debug       string
)

func init() {
	flag.StringVar(&configFile, "config", "config.json", "The path to the JSON-structured configuration file")
	flag.StringVar(&gcpAuthTokenFile, "gcptoken", "client_secret_XXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX.apps.googleusercontent.com.json", "The path to the JSON-structured Google Cloud Platform authentication token")
	flag.StringVar(&configIsBot, "bot", "false", "Whether or not to act as a bot")
	flag.IntVar(&masterPID, "masterpid", -1, "The bot master's PID")
	flag.StringVar(&killOldBot, "killold", "false", "Whether or not to kill an old bot process")
	flag.StringVar(&debug, "debug", "false", "Whether or not to output debugging and trace messages")
	flag.Parse()

	if configIsBot == "true" {
		logFile, err := os.OpenFile("clinet.bot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic("Error creating log file: " + err.Error())
		}
		initLogging(logFile, "BOT", debug)
	} else {
		logFile, err := os.OpenFile("clinet.main.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic("Error creating log file: " + err.Error())
		}
		initLogging(logFile, "MAIN", debug)
	}
}

func main() {
	defer recoverPanic()
	defer logFile.Close()

	Info.Println("Clinet © JoshuaDoes: 2017-2021.")
	Info.Println("Build ID: " + BuildID)
	Info.Println("Current PID: " + strconv.Itoa(os.Getpid()))

	uptime = time.Now()

	if configIsBot == "true" {
		numCPU := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPU * 2)

		Debug.Printf("CPU Core Count: %d\n", numCPU)
		Debug.Printf("Max Process Count: %d\n", numCPU*2)

		Info.Println("Loading settings...")
		configFileHandle, err := os.Open(configFile)
		defer configFileHandle.Close()
		if err != nil {
			Error.Println(err)
			os.Exit(1)
		} else {
			configParser := json.NewDecoder(configFileHandle)
			if err = configParser.Decode(&botData); err != nil {
				Error.Println(err)
				os.Exit(1)
			} else {
				configErr := botData.PrepConfig() //Check the configuration for any errors or inconsistencies, then prepare it for usage
				if configErr != nil {
					Error.Println(configErr)
					os.Exit(1)
				}
			}
		}

		Info.Println("Initializing clients for external services...")
		if gcpAuthTokenFile != "" {
			tokenJSON, err := ioutil.ReadFile(gcpAuthTokenFile)
			if err != nil {
				Error.Printf("Error initializing Google Assistant: %v", err)
			} else {
				gcpToken = &gassist.Token{}
				err = json.Unmarshal(tokenJSON, &gcpToken)
				if err != nil {
					Error.Printf("Error initializing Google Assistant: %v", err)
				} else {
					googleAssistant, err := gassist.NewAssistant(gcpToken, nil, "en-US", gassist.NewDevice("254636LIVE0001", "assistant-for-clinet"), gassist.NewAudioSettings(1, 1, 16000, 16000, 100))
					if err != nil {
						Error.Printf("Error initializing Google Assistant: %v", err)
					} else {
						botData.BotClients.GoogleAssistant = googleAssistant

						if googleAssistant.GetAuthURL() != "" {
							Warning.Println("Please open the following URL to authenticate with Google Cloud Platform:", googleAssistant.GetAuthURL())
							Warning.Println("When you've authenticated successfully, press enter to continue.")
							bufio.NewReader(os.Stdin).ReadLine()
						}
					}
				}
			}
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
			botData.BotClients.SoundCloud = &soundcloud.Client{ClientID: botData.BotKeys.SoundCloudClientID}
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
				Error.Printf("Error initializing YouTube: %v", err)
			} else {
				botData.BotClients.YouTube = youtubeClient
			}
			botData.BotClients.YTDL = &ytdl.Client{
				Debug: true,
			//	HTTPClient: httpClient,
			}
		}
		if botData.BotOptions.UseGitHub {
			botData.BotClients.GitHub = github.NewClient(nil)
		}
		if botData.BotOptions.UseLyrics {
			botData.BotClients.Lyrics = lyrics.New(lyrics.WithoutProviders(), lyrics.WithLyricsWikia(), lyrics.WithMusixMatch(), lyrics.WithSongLyrics(), lyrics.WithGeniusLyrics(botData.BotKeys.GeniusAccessToken))
		}
		if botData.BotOptions.UseNinty {
			nintyClient, err := fennel.NewAccountServerClient("https://account.nintendo.net/v1/api", ctrCommonCert, ctrCommonKey, botData.BotKeys.Ninty)
			if err != nil {
				Error.Printf("Error initializing Ninty: %v", err)
			} else {
				botData.BotClients.Ninty = nintyClient
			}
		}
		if botData.BotOptions.UseFeed {
			botData.BotClients.FeedParser = gofeed.NewParser()
		}

		Info.Println("Creating a Discord session...")
		discord, err := discordgo.New("Bot " + botData.BotToken)
		if err != nil {
			panic(err)
		}
		if debug == "true" {
			discord.LogLevel = discordgo.LogInformational
		}

		Info.Println("Registering Discord event handlers...")
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

		//If a state exists, load it
		Info.Println("Loading state...")
		stateRestoreAll()

		Info.Println("Connecting to Discord...")
		err = discord.Open()
		if err != nil {
			panic(err)
		}
		Info.Println("Connected successfully!")
		botData.DiscordSession = discord

		if botData.SendOwnerStackTraces {
			checkPanicRecovery()
		}

		Debug.Println("Checking for first run...")
		if firstRun() {
			ioutil.WriteFile(".firstrun", make([]byte, 0), 0644)
		}

		Debug.Println("Checking if bot was restarted...")
		checkRestart()

		Debug.Println("Checking if bot was updated...")
		checkUpdate()

		if botData.BotOptions.API.Enabled {
			Info.Printf("Starting API on [%s]...\n", botData.BotOptions.API.Host)
			go StartAPI(botData.BotOptions.API.Host)
		}

		Debug.Println("Waiting for SIGINT syscall signal...")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT)
		<-sc

		//Save the current state before shutting down
		// Note: This is done before shutting down as the shutdown process may yield
		//       some errors with goroutines like voice playback
		stateSaveAll()

		botData.BotClients.GoogleAssistant.Close()

		//Leave all voice channels
		for _, voiceIDRow := range voiceData {
			if voiceIDRow.IsConnected() {
				if voiceIDRow.IsStreaming() {
					//Notify users that an update is occuring
					botData.DiscordSession.ChannelMessageSendEmbed(voiceIDRow.TextChannelID, NewEmbed().SetTitle("Update").SetDescription("Your audio playback has been interrupted for a "+botData.BotName+" update event. You may resume playback in a few seconds.").SetColor(0x1C1C1C).MessageEmbed)

					debugLog("> Stopping stream in voice channel "+voiceIDRow.VoiceConnection.ChannelID+"...", false)
					voiceIDRow.Stop()
				}
				debugLog("> Closing connection to voice channel "+voiceIDRow.VoiceConnection.ChannelID+"...", false)
				voiceIDRow.VoiceConnection.Close()
			}
		}

		Info.Println("Disconnecting from Discord...")
		discord.Close()
	} else {
		botPid := spawnBot()
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT)
		watchdogTicker := time.Tick(1 * time.Second)

		for {
			select {
			case _, ok := <-sc:
				if ok {
					botProcess, _ := os.FindProcess(botPid)
					_ = botProcess.Signal(syscall.SIGINT)
					waitProcess(botPid)
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
	if isReady {
		return //We don't want to re-init if we have to reconnect to Discord
	}
	defer recoverPanic()

	Debug.Println("Setting bot username from Discord state...")
	botData.BotName = session.State.User.Username

	Debug.Println("Initializing commands...")
	initCommands()

	Debug.Println("Initializing natural language commands...")
	initNLPCommands()

	Debug.Println("Initializing query service handlers...")
	initQueryServices()

	Debug.Println("Initializing voice service handlers...")
	initVoiceServices()

	Debug.Println("Setting random presence...")
	updateRandomStatus(session, 0)

	Debug.Println("Creating cronjob handler...")
	cronjob := cron.New()

	Debug.Println("Creating random presence update cronjob...")
	cronjob.AddFunc("@every 1m", func() { updateRandomStatus(session, 0) })

	Debug.Println("Creating random tip message cronjob...")
	cronjob.AddFunc("@every 1h", func() { sendTipMessages() })

	Debug.Println("Starting cronjobs...")
	cronjob.Start()

	Debug.Println("Loading active reminders...")
	oldRemindEntries := remindEntries
	remindEntries = make([]RemindEntry, 0)
	for i := range oldRemindEntries {
		remindWhen(oldRemindEntries[i].UserID, oldRemindEntries[i].GuildID, oldRemindEntries[i].ChannelID, oldRemindEntries[i].Message, oldRemindEntries[i].Added, oldRemindEntries[i].When, time.Now())
	}

	Debug.Println("Loading feeds...")
	for guildID, guild := range guildSettings {
		oldFeeds := guild.Feeds
		guild.Feeds = make([]*Feed, 0)
		for _, feed := range oldFeeds {
			addErr := addFeed(guildID, feed.ChannelID, feed.FeedURL, feed.Frequency)
			if addErr != nil {
				Error.Println("Error adding feed ["+feed.FeedLink+"]: %v", addErr)
			}
		}
	}

	if gcpAuthURL := botData.BotClients.GoogleAssistant.GetAuthURL(); gcpAuthURL != "" {
		ownerPrivChannel, err := botData.DiscordSession.UserChannelCreate(botData.BotOwnerID)
		if err != nil {
			debugLog("An error occurred creating a private channel with the bot owner.", false)
		} else {
			ownerPrivChannelID := ownerPrivChannel.ID
			botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, "Authenticate with the Google Assistant: "+gcpAuthURL)
		}
	}

	isReady = true
	Info.Println("Discord is ready!")
}

func updateRandomStatus(session *discordgo.Session, status int) {
	if status == 0 {
		status = rand.Intn(len(botData.CustomStatuses)) + 1
	}
	status--

	session.UpdateStatusComplex(discordgo.UpdateStatusData{Activities: []*discordgo.Activity{botData.CustomStatuses[status]}})
	Debug.Printf("Presence: ", botData.CustomStatuses[status])
}

func updateListeningStatus(session *discordgo.Session, artist, title string) {
	session.UpdateListeningStatus(artist + " - " + title)
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

func typingEvent(session *discordgo.Session, channelID string, updatedMessageEvent bool) {
	if botData.BotOptions.SendTypingEvent && updatedMessageEvent == false {
		Debug.Printf("Typing in channel %s...\n", channelID)
		session.ChannelTyping(channelID)
	}
}

func debugLog(msg string, overrideConfig bool) {
	if botData.DebugMode || overrideConfig {
		Debug.Println(msg)
	}
}

func stateSaveAll() {
	if _, err := os.Stat("state"); os.IsNotExist(err) {
		os.Mkdir("state", 0744)
	}

	err := stateSaveRaw(guildData, "state/guildData.json")
	if err != nil {
		Error.Printf("Error saving guildData state: %s\n", err)
	}

	err = stateSaveRaw(guildSettings, "state/guildSettings.json")
	if err != nil {
		Error.Printf("Error saving guildSettings state: %s\n", err)
	}

	err = stateSaveRaw(userSettings, "state/userSettings.json")
	if err != nil {
		Error.Printf("Error saving userSettings state: %s\n", err)
	}

	err = stateSaveRaw(starboards, "state/starboards.json")
	if err != nil {
		Error.Printf("Error saving starboards: %s\n", err)
	}

	err = stateSaveRaw(remindEntries, "state/reminds.json")
	if err != nil {
		Error.Printf("Error saving reminders: %s\n", err)
	}

	err = stateSaveRaw(voiceData, "state/voiceData.json")
	if err != nil {
		Error.Printf("Error saving voiceData state: %s\n", err)
	}
}

func stateSaveRaw(data interface{}, file string) error {
	dataJSON, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, dataJSON, 0744)
	return err
}

func stateRestoreAll() {
	err := stateRestoreRaw("state/guildData.json", &guildData)
	if err != nil {
		Error.Printf("Error loading guildData state: %s\n", err)
	}

	err = stateRestoreRaw("state/guildSettings.json", &guildSettings)
	if err != nil {
		Error.Printf("Error loading guildSettings state: %s\n", err)
	}

	err = stateRestoreRaw("state/userSettings.json", &userSettings)
	if err != nil {
		Error.Printf("Error loading userSettings state: %s\n", err)
	}

	err = stateRestoreRaw("state/starboards.json", &starboards)
	if err != nil {
		Error.Printf("Error loading starboards: %s\n", err)
	}

	err = stateRestoreRaw("state/reminds.json", &remindEntries)
	if err != nil {
		Error.Printf("Error loading reminders: %s\n", err)
	}

	err = stateRestoreRaw("state/voiceData.json", &voiceData)
	if err != nil {
		Error.Printf("Error loading voiceData state: %s\n", err)
	}
}

func stateRestoreRaw(file string, data interface{}) error {
	dataJSON, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal(dataJSON, data)
	return err
}

func firstRun() bool {
	_, err := ioutil.ReadFile(".firstrun")
	if err == nil {
		DowntimeReason = "Restarted by host system or <@" + botData.BotOwnerID + ">"
		return false
	}

	DowntimeReason = "First run"
	return true
}

func checkRestart() {
	restartChannelID, err := ioutil.ReadFile(".restart")
	if err == nil && len(restartChannelID) > 0 {
		DowntimeReason = "Restarted by <@" + botData.BotOwnerID + ">"

		Info.Println("Restart succeeded!")
		restartEmbed := NewGenericEmbed("Restart", "Successfully restarted "+botData.BotName+"!")
		botData.DiscordSession.ChannelMessageSendEmbed(string(restartChannelID), restartEmbed)

		os.Remove(".restart")
	}
}

func checkUpdate() {
	updateChannelID, err := ioutil.ReadFile(".update")
	if err == nil && len(updateChannelID) > 0 {
		DowntimeReason = "Updated to " + BuildID

		Info.Println("Update succeeded!")
		updateEmbed := NewGenericEmbed("Update", "Successfully updated "+botData.BotName+"!")
		botData.DiscordSession.ChannelMessageSendEmbed(string(updateChannelID), updateEmbed)

		os.Remove(".update")
	}
}
