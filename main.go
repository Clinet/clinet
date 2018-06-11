package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JoshuaDoes/duckduckgolang"      //Allows the usage of the DuckDuckGo API
	"github.com/JoshuaDoes/go-soundcloud"       //Allows usage of the SoundCloud API
	"github.com/JoshuaDoes/go-wolfram"          //Allows usage of the Wolfram|Alpha API
	"github.com/bwmarrin/discordgo"             //Allows usage of the Discord API
	"github.com/google/go-github/github"        //Allows usage of the GitHub API
	"github.com/jonas747/dca"                   //Allows the encoding/decoding of the Discord Audio format
	"github.com/koffeinsource/go-imgur"         //Allows usage of the Imgur API
	"github.com/koffeinsource/go-klogger"       //For some reason, this is required for go-imgur's logging
	"github.com/nishanths/go-xkcd"              //Allows the fetching of XKCD comics
	"github.com/paked/configure"                //Allows configuration of the program via external sources
	"github.com/robfig/cron"                    //Allows for better management of running tasks at specific intervals
	"github.com/rylio/ytdl"                     //Allows the fetching of YouTube video metadata and download URLs
	"google.golang.org/api/googleapi/transport" //Allows the making of authenticated API requests to Google
	"google.golang.org/api/youtube/v3"          //Allows usage of the YouTube API
)

//Bot data structs
type BotClients struct {
	DuckDuckGo *duckduckgo.Client
	GitHub     *github.Client
	Imgur      imgur.Client
	SoundCloud *soundcloud.Client
	Wolfram    *wolfram.Client
	XKCD       *xkcd.Client
	YouTube    *youtube.Service
}
type BotData struct {
	BotClients      BotClients
	BotKeys         BotKeys               `json:"botKeys"`
	BotName         string                `json:"botName"`
	BotOptions      BotOptions            `json:"botOptions"`
	BotToken        string                `json:"botToken"`
	CommandPrefix   string                `json:"cmdPrefix"`
	CustomResponses []CustomResponseQuery `json:"customResponses"`
	CustomStatuses  []CustomStatus        `json:"customStatuses"`
	DebugMode       bool                  `json:"debugMode"`
}
type BotKeys struct {
	DuckDuckGoAppName    string `json:"ddgAppName"`
	ImgurClientID        string `json:"imgurClientID"`
	SoundCloudAppVersion string `json:"soundcloudAppVersion"`
	SoundCloudClientID   string `json:"soundcloudClientID"`
	WolframAppID         string `json:"wolframAppID"`
	YouTubeAPIKey        string `json:"youtubeAPIKey"`
}
type BotOptions struct {
	SendTypingEvent   bool     `json:"sendTypingEvent"`
	UseDuckDuckGo     bool     `json:"useDuckDuckGo"`
	UseGitHub         bool     `json:"useGitHub"`
	UseImgur          bool     `json:"useImgur"`
	UseSoundCloud     bool     `json:"useSoundCloud"`
	UseWolframAlpha   bool     `json:"useWolframAlpha"`
	UseXKCD           bool     `json:"useXKCD"`
	UseYouTube        bool     `json:"useYouTube"`
	WolframDeniedPods []string `json:"wolframDeniedPods"`
}
type CustomResponseQuery struct {
	Expression string `json:"expression"`
	Regexp     *regexp.Regexp
	Responses  []CustomResponseReply `json:"response"`
}
type CustomResponseReply struct {
	Response string `json:"text"`
	ImageURL string `json:"imageURL"`
}
type CustomStatus struct {
	Type   int    `json:"type"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

func (configData *BotData) PrepConfig() error {
	//Bot config checks
	if configData.BotName == "" {
		return errors.New("config:{botName: \"\"}")
	}
	if configData.BotToken == "" {
		return errors.New("config:{botName: \"\"}")
	}
	if configData.CommandPrefix == "" {
		return errors.New("config:{cmdPrefix: \"\"}")
	}

	//Bot key checks
	if configData.BotOptions.UseDuckDuckGo && configData.BotKeys.DuckDuckGoAppName == "" {
		return errors.New("config:{botOptions:{useDuckDuckGo: true}} not permitted, config:{botKeys:{ddgAppName: \"\"}}")
	}
	if configData.BotOptions.UseImgur && configData.BotKeys.ImgurClientID == "" {
		return errors.New("config:{botOptions:{useImgur: true}} not permitted, config:{botKeys:{imgurClientID: \"\"}}")
	}
	if configData.BotOptions.UseSoundCloud && configData.BotKeys.SoundCloudAppVersion == "" {
		return errors.New("config:{botOptions:{useSoundCloud: true}} not permitted, config:{botKeys:{soundcloudAppVersion: \"\"}}")
	}
	if configData.BotOptions.UseSoundCloud && configData.BotKeys.SoundCloudClientID == "" {
		return errors.New("config:{botOptions:{useSoundCloud: true}} not permitted, config:{botKeys:{soundcloudClientID: \"\"}}")
	}
	if configData.BotOptions.UseWolframAlpha && configData.BotKeys.WolframAppID == "" {
		return errors.New("config:{botOptions:{useWolframAlpha: true}} not permitted, config:{botKeys:{wolframAppID: \"\"}}")
	}
	if configData.BotOptions.UseYouTube && configData.BotKeys.YouTubeAPIKey == "" {
		return errors.New("config:{botOptions:{useYouTube: true}} not permitted, config:{botKeys:{youtubeAPIKey: \"\"}}")
	}

	//Custom response checks
	for i, customResponse := range configData.CustomResponses {
		regexp, err := regexp.Compile(customResponse.Expression)
		if err != nil {
			return err
		} else {
			configData.CustomResponses[i].Regexp = regexp
		}
	}
	return nil
}

//Guild data structs
type GuildData struct {
	AudioQueue      []AudioQueueEntry
	AudioNowPlaying AudioQueueEntry
	VoiceData       VoiceData
	Queries         map[string]*Query
	YouTubeResults  map[string]*YouTubeResultNav
}

func (guild *GuildData) QueueAddData(author, imageURL, title, thumbnailURL, mediaURL, sourceType string, requester *discordgo.User) {
	var queueData AudioQueueEntry
	queueData.Author = author
	queueData.ImageURL = imageURL
	queueData.MediaURL = mediaURL
	queueData.Requester = requester
	queueData.ThumbnailURL = thumbnailURL
	queueData.Title = title
	queueData.Type = sourceType
	guild.AudioQueue = append(guild.AudioQueue, queueData)
}
func (guild *GuildData) QueueAdd(audioQueueEntry AudioQueueEntry) {
	guild.AudioQueue = append(guild.AudioQueue, audioQueueEntry)
}
func (guild *GuildData) QueueRemove(entry int) {
	guild.AudioQueue = append(guild.AudioQueue[:entry], guild.AudioQueue[entry+1:]...)
}
func (guild *GuildData) QueueRemoveRange(start int, end int) {
	for entry := end; entry < start; entry-- {
		guild.AudioQueue = append(guild.AudioQueue[:entry], guild.AudioQueue[entry+1:]...)
	}
}
func (guild *GuildData) QueueClear() {
	guild.AudioQueue = nil
}
func (guild *GuildData) QueueGet(guildID string, entry int) AudioQueueEntry {
	if len(guildData[guildID].AudioQueue) >= entry {
		return guildData[guildID].AudioQueue[entry]
	} else {
		return AudioQueueEntry{}
	}
}
func (guild *GuildData) QueueGetNext(guildID string) AudioQueueEntry {
	if len(guildData[guildID].AudioQueue) > 0 {
		return guildData[guildID].AudioQueue[0]
	} else {
		return AudioQueueEntry{}
	}
}

//YouTube search results, interacted with via commands
type YouTubeResultNav struct {
	//Used by struct functions
	Query         string                  //The search query used to retrieve the current results
	TotalResults  int64                   //The total amount of results for the current search query
	Results       []*youtube.SearchResult //The results of the current page
	PrevPageToken string                  //The token of the previous page of results
	NextPageToken string                  //The token of the next page of results
	PageNumber    int                     //The numerical identifier of the current page

	//Used by external functions for easy page management
	ResponseID string //The message response ID used to display and update result listings
	MaxResults int64  //The total amount of results per page
}

func (page *YouTubeResultNav) Prev() error {
	if page.PageNumber == 0 {
		return errors.New("No search pages found")
	}
	if page.PrevPageToken == "" {
		return errors.New("No pages exist before current page")
	}

	searchCall := botData.BotClients.YouTube.Search.
		List("id").
		Q(page.Query).
		PageToken(page.PrevPageToken)

	response, err := searchCall.Do()
	if err != nil {
		return errors.New("Could not find any video results for the previous page")
	}

	page.PageNumber -= 1
	page.Results = response.Items
	page.PrevPageToken = response.PrevPageToken
	page.NextPageToken = response.NextPageToken

	return nil
}
func (page *YouTubeResultNav) Next() error {
	if page.PageNumber == 0 {
		return errors.New("No search pages found")
	}
	if page.NextPageToken == "" {
		return errors.New("No pages exist after current page")
	}

	searchCall := botData.BotClients.YouTube.Search.
		List("id").
		Q(page.Query).
		PageToken(page.NextPageToken)

	response, err := searchCall.Do()
	if err != nil {
		return errors.New("Could not find any video results for the next page")
	}

	page.PageNumber += 1
	page.Results = response.Items
	page.PrevPageToken = response.PrevPageToken
	page.NextPageToken = response.NextPageToken

	return nil
}
func (page *YouTubeResultNav) GetResults() ([]*youtube.SearchResult, error) {
	if len(page.Results) == 0 {
		return nil, errors.New("No search results found")
	}
	return page.Results, nil
}
func (page *YouTubeResultNav) Search(query string) error {
	if page.MaxResults == 0 {
		page.MaxResults = 5
	}

	page.Query = ""
	page.PageNumber = 0
	page.TotalResults = 0
	page.Results = nil
	page.PrevPageToken = ""
	page.NextPageToken = ""

	searchCall := botData.BotClients.YouTube.Search.
		List("id").
		Q(query).
		MaxResults(page.MaxResults).
		Type("video")

	response, err := searchCall.Do()
	if err != nil {
		return errors.New("Could not find any video results for the specified query")
	}

	page.Query = query
	page.PageNumber = 1
	page.TotalResults = response.PageInfo.TotalResults
	page.Results = response.Items
	page.PrevPageToken = response.PrevPageToken
	page.NextPageToken = response.NextPageToken

	return nil
}

type DynamicSettings struct {
	Guilds []GuildSettings `json:"guilds"` //An array of guild IDs with settings for each guild
	Users  []UserSettings  `json:"users"`  //An array of user IDs with settings for each user
}
type GuildSettings struct { //By default this will only be configurable for users in a role with the server admin permission
	AllowVoice              bool                  `json:"allowVoice"`              //Whether voice commands should be usable in this guild
	BotAdminRoles           []string              `json:"adminRoles"`              //An array of role IDs that can admin the bot
	BotAdminUsers           []string              `json:"adminUsers"`              //An array of user IDs that can admin the bot
	BotName                 string                `json:"botName"`                 //The bot name to use in this guild
	BotOptions              BotOptions            `json:"botOptions"`              //The bot options to use in this guild (true gets overridden if global bot config is false)
	BotPrefix               string                `json:"botPrefix"`               //The bot prefix to use in this guild
	CustomResponses         []CustomResponseQuery `json:"customResponses"`         //An array of custom responses specific to the guild
	UserJoinMessage         string                `json:"userJoinMessage"`         //A message to send when a user joins
	UserJoinMessageChannel  string                `json:"userJoinMessageChannel"`  //The channel to send the user join message to
	UserLeaveMessage        string                `json:"userLeaveMessage"`        //A message to send when a user leaves
	UserLeaveMessageChannel string                `json:"userLeaveMessageChannel"` //The channel to send the user leave message to
}
type UserSettings struct {
	Balance     int64  `json:"balance"`     //A balance to use as virtual currency for some bot tasks
	Description string `json:"description"` //A description set by the user
}

type AudioQueueEntry struct {
	Author           string
	ImageURL         string
	MediaURL         string
	Requester        *discordgo.User
	RequestMessageID string //Used for if someone edits their request
	ThumbnailURL     string
	Title            string
	Type             string
}

func (audioQueueEntry *AudioQueueEntry) GetNowPlayingEmbed() *discordgo.MessageEmbed {
	switch audioQueueEntry.Type {
	case "youtube":
		return NewEmbed().
			SetTitle("Now Playing from YouTube").
			AddField(audioQueueEntry.Title, audioQueueEntry.Author).
			AddField("Requester", audioQueueEntry.Requester.String()).
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF0000).MessageEmbed
	case "soundcloud":
		return NewEmbed().
			SetTitle("Now Playing from SoundCloud").
			AddField(audioQueueEntry.Title, audioQueueEntry.Author).
			AddField("Requester", audioQueueEntry.Requester.String()).
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF7700).MessageEmbed
	default:
		return NewEmbed().
			SetTitle("Now Playing").
			AddField("URL", audioQueueEntry.MediaURL).
			AddField("Requester", audioQueueEntry.Requester.String()).
			SetColor(0x1C1C1C).MessageEmbed
	}
}
func (audioQueueEntry *AudioQueueEntry) GetNowPlayingDurationEmbed(stream *dca.StreamingSession) *discordgo.MessageEmbed {
	//Get the current duration
	currentDuration := secondsToHuman(stream.PlaybackPosition().Seconds())

	switch audioQueueEntry.Type {
	case "youtube":
		return NewEmbed().
			SetTitle("Now Playing from YouTube").
			AddField(audioQueueEntry.Title, audioQueueEntry.Author).
			AddField("Requester", audioQueueEntry.Requester.String()).
			AddField("Duration", currentDuration).
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF0000).MessageEmbed
	case "soundcloud":
		return NewEmbed().
			SetTitle("Now Playing from SoundCloud").
			AddField(audioQueueEntry.Title, audioQueueEntry.Author).
			AddField("Requester", audioQueueEntry.Requester.String()).
			AddField("Duration", currentDuration).
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF7700).MessageEmbed
	default:
		return NewEmbed().
			SetTitle("Now Playing").
			AddField("URL", audioQueueEntry.MediaURL).
			AddField("Requester", audioQueueEntry.Requester.String()).
			AddField("Duration", currentDuration).
			SetColor(0x1C1C1C).MessageEmbed
	}
}
func (audioQueueEntry *AudioQueueEntry) GetQueueAddedEmbed() *discordgo.MessageEmbed {
	if audioQueueEntry.Type == "" {
		if isYouTubeURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "youtube"
		} else if isSoundCloudURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "soundcloud"
		} else {
			audioQueueEntry.Type = "direct"
		}
	}

	switch audioQueueEntry.Type {
	case "youtube":
		return NewEmbed().
			SetTitle("Added to Queue from YouTube").
			AddField(audioQueueEntry.Title, audioQueueEntry.Author).
			AddField("Requester", audioQueueEntry.Requester.String()).
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF0000).MessageEmbed
	case "soundcloud":
		return NewEmbed().
			SetTitle("Added to Queue from SoundCloud").
			AddField(audioQueueEntry.Title, audioQueueEntry.Author).
			AddField("Requester", audioQueueEntry.Requester.String()).
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF7700).MessageEmbed
	default:
		return NewEmbed().
			SetTitle("Added to Queue").
			AddField("URL", audioQueueEntry.MediaURL).
			AddField("Requester", audioQueueEntry.Requester.String()).
			SetColor(0x1C1C1C).MessageEmbed
	}
}
func (audioQueueEntry *AudioQueueEntry) FillMetadata() {
	if audioQueueEntry.Type == "" {
		if isYouTubeURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "youtube"
		} else if isSoundCloudURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "soundcloud"
		} else {
			audioQueueEntry.Type = "direct"
		}
	}

	switch audioQueueEntry.Type {
	case "youtube":
		videoInfo, err := ytdl.GetVideoInfo(audioQueueEntry.MediaURL)
		if err != nil {
			return
		}
		audioQueueEntry.Author = videoInfo.Author
		audioQueueEntry.ImageURL = videoInfo.GetThumbnailURL("maxresdefault").String()
		audioQueueEntry.ThumbnailURL = videoInfo.GetThumbnailURL("default").String()
		audioQueueEntry.Title = videoInfo.Title
	case "soundcloud":
		audioInfo, err := botData.BotClients.SoundCloud.GetTrackInfo(audioQueueEntry.MediaURL)
		if err != nil {
			return
		}
		audioQueueEntry.Author = audioInfo.Artist
		audioQueueEntry.ImageURL = audioInfo.ArtURL
		audioQueueEntry.ThumbnailURL = audioInfo.ArtURL
		audioQueueEntry.Title = audioInfo.Title
	}
}

type Query struct {
	ResponseMessageID string
}
type VoiceData struct {
	VoiceConnection     *discordgo.VoiceConnection
	EncodingSession     *dca.EncodeSession
	StreamingSession    *dca.StreamingSession
	ChannelIDJoinedFrom string

	IsPlaybackPreparing bool //Whether or not the playback is being prepared
	IsPlaybackRunning   bool //Whether or not playback is currently running
	WasStoppedManually  bool //Whether or not playback was stopped manually or automatically
	WasSkipped          bool //Whether or not playback was skipped

	//Configuration settings that can be set via commands
	RepeatLevel int //0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing
}

var (
	//Variables filled in on compile time using github.com/JoshuaDoes/govvv
	GitBranch     string
	GitCommit     string
	GitCommitFull string
	GitCommitMsg  string
	GitState      string
	BuildDate     string

	//A unique build ID inspired by the Android Open Source Project
	BuildID string = "clinet_discord-" + GitState + " " + GitBranch + "-" + GitCommit

	//The URL to the current commit
	GitHubCommitURL string = "https://github.com/JoshuaDoes/clinet-discord/commit/" + GitCommitFull

	//The version of Go used to build this release
	GolangVersion string = runtime.Version()
)

var (
	botData   *BotData = &BotData{}
	guildData          = make(map[string]*GuildData)

	conf                  = configure.New()
	confConfigFile        = conf.String("config", "config.json", "The location of the JSON-structured configuration file")
	configFile     string = ""
)

func init() {
	conf.Use(configure.NewFlag())
}

func main() {
	debugLog("Clinet-Discord © JoshuaDoes: 2018.\n", true)

	debugLog("> Loading settings...", true)
	conf.Parse()
	configFile = *confConfigFile
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

	debugLog("> Creating a Discord session...", true)
	discord, err := discordgo.New("Bot " + botData.BotToken)
	if err != nil {
		panic(err)
	}

	debugLog("> Registering event handlers...", false)
	discord.AddHandler(discordMessageCreate)
	discord.AddHandler(discordMessageDelete)
	discord.AddHandler(discordMessageDeleteBulk)
	discord.AddHandler(discordMessageUpdate)
	discord.AddHandler(discordReady)

	debugLog("> Connecting to Discord...", true)
	err = discord.Open()
	if err != nil {
		panic(err)
	}
	debugLog("> Connection successful", true)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	for _, guildDataRow := range guildData {
		if guildDataRow.VoiceData.VoiceConnection != nil {
			debugLog("> Closing connection to voice channel "+guildDataRow.VoiceData.VoiceConnection.ChannelID+"...", false)
			guildDataRow.VoiceData.VoiceConnection.Close()
		}
	}
	debugLog("> Disconnecting from Discord...", true)
	discord.Close()
}

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: "+fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}

	go handleMessage(session, message, false)
}
func discordMessageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: "+fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}

	go handleMessage(session, message, true)
}
func discordMessageDelete(session *discordgo.Session, event *discordgo.MessageDelete) {
	message := event //Make it easier to keep track of what's happening

	debugLog("[D] ID: "+message.ID, false)

	guildChannel, err := session.Channel(message.ChannelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			_, messageFound := guildData[guildID].Queries[message.ID]
			if messageFound {
				debugLog("> Deleting message...", false)
				session.ChannelMessageDelete(message.ChannelID, guildData[guildID].Queries[message.ID].ResponseMessageID) //Delete the query response message
				guildData[guildID].Queries[message.ID] = nil                                                              //Remove the message from the query list
			} else {
				debugLog("> Error finding deleted message in queries list", false)
			}
		} else {
			debugLog("> Error finding guild for deleted message", false)
		}
	} else {
		debugLog("> Error finding channel for deleted message", false)
	}
}
func discordMessageDeleteBulk(session *discordgo.Session, event *discordgo.MessageDeleteBulk) {
	messages := event.Messages
	channelID := event.ChannelID

	guildChannel, err := session.Channel(channelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			for i := 0; i > len(messages); i++ {
				debugLog("[D] ID: "+messages[i], false)
				_, messageFound := guildData[guildID].Queries[messages[i]]
				if messageFound {
					debugLog("> Deleting message...", false)
					session.ChannelMessageDelete(channelID, guildData[guildID].Queries[messages[i]].ResponseMessageID) //Delete the query response message
					guildData[guildID].Queries[messages[i]] = nil                                                      //Remove the message from the query list
				} else {
					debugLog("> Error finding deleted message in queries list", false)
				}
			}
		}
	}
}
func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	updateRandomStatus(session, 0)
	cronjob := cron.New()
	cronjob.AddFunc("@every 1m", func() { updateRandomStatus(session, 0) })
	cronjob.Start()
}

func handleMessage(session *discordgo.Session, message *discordgo.Message, updatedMessageEvent bool) {
	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		debugLog("> Error finding message channel", false)
		return //Error finding the channel
	}
	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		debugLog("> Error finding message guild", false)
		return //Error finding the guild
	}
	content := message.Content
	if content == "" {
		return //The message was empty
	}
	contentReplaced, _ := message.ContentWithMoreMentionsReplaced(session)

	/*
		//If message is single-lined
			[New][District JD - #main] @JoshuaDoes#0001: Hello, world!

		//If message is multi-lined
			[New][District JD - #main] @JoshuaDoes#0001:
			Hello, world!
			My name is Joshua.
			This is a lot of fun!

		//If user is bot
			[New][District JD - #main] *Clinet#1823: Hello, world!
	*/
	eventType := "[New]"
	if updatedMessageEvent {
		eventType = "[Updated]"
	}
	userType := "@"
	if message.Author.Bot {
		userType = "*"
	}
	if strings.Contains(content, "\n") {
		debugLog(eventType+"["+guild.Name+" - #"+channel.Name+"] "+userType+message.Author.Username+"#"+message.Author.Discriminator+":\n"+contentReplaced, false)
	} else {
		debugLog(eventType+"["+guild.Name+" - #"+channel.Name+"] "+userType+message.Author.Username+"#"+message.Author.Discriminator+": "+contentReplaced, false)
	}

	var responseEmbed *discordgo.MessageEmbed

	if strings.HasPrefix(content, botData.CommandPrefix) {
		cmdMsg := strings.Replace(content, botData.CommandPrefix, "", 1)
		cmd := strings.Split(cmdMsg, " ")

		switch cmd[0] {
		case "help":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - Help").
				SetDescription("A list of available commands for "+botData.BotName+".").
				AddField(botData.CommandPrefix+"help", "Displays this help message.").
				AddField(botData.CommandPrefix+"about", "Displays information about "+botData.BotName+" and how to use it.").
				AddField(botData.CommandPrefix+"version", "Displays the current version of "+botData.BotName+".").
				AddField(botData.CommandPrefix+"credits", "Displays a list of credits for the creation and functionality of "+botData.BotName+".").
				AddField(botData.CommandPrefix+"roll", "Rolls a dice.").
				AddField(botData.CommandPrefix+"doubleroll", "Rolls two die.").
				AddField(botData.CommandPrefix+"coinflip", "Flips a coin.").
				AddField(botData.CommandPrefix+"xkcd (comic number|random|latest)", "Displays an xkcd comic depending on the requested type or comic number.").
				AddField(botData.CommandPrefix+"imgur (url)", "Displays info about the specified Imgur image, album, gallery image, or gallery album.").
				AddField(botData.CommandPrefix+"github/gh username(/repo_name)", "Displays info about the specified GitHub user or repo.").
				AddField(botData.CommandPrefix+"play (url/YouTube search query)", "Plays either the first result from the specified YouTube search query or the specified YouTube/direct audio URL in the user's current voice channel.").
				AddField(botData.CommandPrefix+"youtube search (query)", "Displays paginated results of the specified YouTube search query with a command list for navigating and selecting a result.").
				AddField(botData.CommandPrefix+"stop", "Stops the currently playing audio.").
				AddField(botData.CommandPrefix+"skip", "Stops the currently playing audio, and, if available, attempts to play the next audio in the queue.").
				AddField(botData.CommandPrefix+"repeat", "Switches the repeat level between the entire guild queue, the currently now playing audio, and not repeating at all.").
				AddField(botData.CommandPrefix+"shuffle", "Shuffles the current guild queue.").
				AddField(botData.CommandPrefix+"queue help", "Lists all available queue commands.").
				AddField(botData.CommandPrefix+"nowplaying/np", "Get info about the currently playing audio.").
				AddField(botData.CommandPrefix+"leave", "Leaves the current voice channel.").
				SetColor(0xFAFAFA).MessageEmbed
		case "about":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - About").
				SetDescription(botData.BotName+" is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
				AddField("How can I use "+botData.BotName+" in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
				AddField("How can I help keep "+botData.BotName+" running?", "The best ways to help keep "+botData.BotName+" running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
				AddField("How can I use "+botData.BotName+"?", "There are many ways to make use of "+botData.BotName+".\n1) Type ``cli$help`` and try using some of the available commands.\n2) Ask "+botData.BotName+" a question, ex: ``"+botData.BotName+", what time is it?`` or ``"+botData.BotName+", what is DiscordApp?``.").
				AddField("Where can I join the "+botData.BotName+" Discord server?", "If you would like to get help and support with "+botData.BotName+" or experiment with the latest and greatest of "+botData.BotName+", use the Discord Server Invite Link at the end of this message.").
				AddField("Invite Link", "https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot").
				AddField("Donation Link", "https://www.paypal.me/JoshuaDoes").
				AddField("Source Code Link", "https://github.com/JoshuaDoes/clinet-discord/").
				AddField("Discord Server Invite Link", "https://discord.gg/qkbKEWT").
				SetColor(0x1C1C1C).MessageEmbed
		case "version":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - Version").
				AddField("Build ID", BuildID).
				AddField("Build Date", BuildDate).
				AddField("Latest Development", GitCommitMsg).
				AddField("GitHub Commit URL", GitHubCommitURL).
				AddField("Golang Version", GolangVersion).
				SetColor(0x1C1C1C).MessageEmbed
		case "credits":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - Credits").
				AddField("Bot Development", "- JoshuaDoes (2018)").
				AddField("Programming Language", "- Golang").
				AddField("Golang Libraries", "- https://github.com/bwmarrin/discordgo\n"+
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
		case "roll":
			random := rand.Intn(6) + 1
			responseEmbed = NewGenericEmbed("Roll", "You rolled a "+strconv.Itoa(random)+"!")
		case "doubleroll", "rolldouble":
			random1 := rand.Intn(6) + 1
			random2 := rand.Intn(6) + 1
			randomTotal := random1 + random2
			responseEmbed = NewGenericEmbed("Double Roll", "You rolled a "+strconv.Itoa(random1)+" and a "+strconv.Itoa(random2)+". The total is "+strconv.Itoa(randomTotal)+"!")
		case "coinflip", "flipcoin":
			random := rand.Intn(2)
			switch random {
			case 0:
				responseEmbed = NewGenericEmbed("Coin Flip", "You got heads!")
			case 1:
				responseEmbed = NewGenericEmbed("Coin Flip", "You got tails!")
			}
		case "imgur":
			if len(cmd) > 1 {
				responseEmbed, err = queryImgur(cmd[1])
				if err != nil {
					responseEmbed = NewErrorEmbed("Imgur Error", fmt.Sprintf("%v", err))
				}
			} else {
				responseEmbed = NewErrorEmbed("Imgur Error", "You must specify an Imgur URL to query Imgur with.")
			}
		case "xkcd":
			if len(cmd) > 1 {
				switch cmd[1] {
				case "random": //Get random XKCD comic
					comic, err := botData.BotClients.XKCD.Random()
					if err != nil {
						responseEmbed = NewErrorEmbed("XKCD Error", "Could not find a random XKCD comic.")
					} else {
						responseEmbed = NewEmbed().
							SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
							SetDescription(comic.Title).
							SetImage(comic.ImageURL).
							SetColor(0x96A8C8).MessageEmbed
					}
				case "latest": //Get latest XKCD comic
					comic, err := botData.BotClients.XKCD.Latest()
					if err != nil {
						responseEmbed = NewErrorEmbed("XKCD Error", "Could not find the latest XKCD comic.")
					} else {
						responseEmbed = NewEmbed().
							SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
							SetDescription(comic.Title).
							SetImage(comic.ImageURL).
							SetColor(0x96A8C8).MessageEmbed
					}
				default: //Get specified XKCD comic
					comicNumber, err := strconv.Atoi(cmd[1])
					if err != nil { //Specified comic is not a valid integer
						responseEmbed = NewErrorEmbed("XKCD Error", "``"+cmd[1]+"`` is not a valid number.")
					} else {
						comic, err := botData.BotClients.XKCD.Get(comicNumber)
						if err != nil {
							responseEmbed = NewErrorEmbed("XKCD Error", "Could not find XKCD comic #"+cmd[1]+".")
						} else {
							responseEmbed = NewEmbed().
								SetTitle("xkcd - #" + cmd[1]).
								SetDescription(comic.Title).
								SetImage(comic.ImageURL).
								SetColor(0x96A8C8).MessageEmbed
						}
					}
				}
			} else { //Get random XKCD comic
				comic, err := botData.BotClients.XKCD.Random()
				if err != nil {
					responseEmbed = NewErrorEmbed("XKCD Error", "Error finding random XKCD comic.")
				} else {
					responseEmbed = NewEmbed().
						SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
						SetDescription(comic.Title).
						SetImage(comic.ImageURL).
						SetColor(0x96A8C8).MessageEmbed
				}
			}
		case "github", "gh":
			// https://godoc.org/github.com/google/go-github/github
			if len(cmd) > 1 {
				request := strings.Split(cmd[1], "/")
				switch len(request) {
				case 1: //A user was specified
					user, err := GitHubFetchUser(request[0])
					if err != nil {
						responseEmbed = NewErrorEmbed("GitHub Error", "There was an error finding info about that user.")
					} else {
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

						for i := 0; i < len(fields); i++ {
							debugLog(fields[i].Name+": "+fields[i].Value, false)
						}

						//Build embed about user
						responseEmbed = NewEmbed().
							SetTitle("GitHub User: " + *user.Login).
							SetImage(*user.AvatarURL).
							SetColor(0x24292D).MessageEmbed
						responseEmbed.Fields = fields
					}
				case 2: //A repo under a user was specified
					repo, err := GitHubFetchRepo(request[0], request[1])
					if err != nil {
						responseEmbed = NewErrorEmbed("GitHub Error", "There was an error finding info about that repo.")
					} else {
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

						for i := 0; i < len(fields); i++ {
							debugLog(fields[i].Name+": "+fields[i].Value, false)
						}

						//Build embed about repo
						responseEmbed = NewEmbed().
							SetTitle("GitHub Repo: " + *repo.FullName).
							SetColor(0x24292D).MessageEmbed
						responseEmbed.Fields = fields
					}
				default:
					responseEmbed = NewErrorEmbed("GitHub Error", "You got a little too specific there! Make sure to only specify either a user or a user/repo combination.")
				}
			} else {
				responseEmbed = NewErrorEmbed("GitHub Error", "You must specify a GitHub user or a GitHub repo to fetch info about.\n\nExamples:\n```"+botData.CommandPrefix+"github JoshuaDoes\n"+botData.CommandPrefix+"gh JoshuaDoes/clinet-discord```")
			}
		case "join":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
					responseEmbed = NewGenericEmbed("Clinet Voice", "Joined voice channel.")
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel to use before using the join command.")
			}
		case "leave":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					if voiceIsStreaming(guild.ID) {
						voiceStop(guild.ID)
					}
					err := voiceLeave(guild.ID, voiceState.ChannelID)
					if err != nil {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error leaving the voice channel.")
					} else {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Left voice channel.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the leave command.")
			}
		case "play":
			if updatedMessageEvent {
				//Todo: Remove this once I figure out how to detect if message update was user-triggered
				//Reason: If you use a YouTube/SoundCloud URL, Discord automatically updates the message with an embed
				//	As far as I know, bots have no way to know if this was a Discord- or user-triggered message update
				//I eventually want users to be able to edit their play command to change a now playing or a queue entry that was misspelled
				return
			}
			if guildData[guild.ID] == nil {
				guildData[guild.ID] = &GuildData{}
				guildData[guild.ID].VoiceData = VoiceData{}
			}
			for guildData[guild.ID].VoiceData.IsPlaybackPreparing {
				//Wait for the handling of a previous playback command to finish
			}
			guildData[guild.ID].VoiceData.IsPlaybackPreparing = true
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
					break
				}
			}
			if foundVoiceChannel {
				if len(cmd) == 1 { //No query or URL was specified
					if voiceIsStreaming(guild.ID) {
						if len(message.Attachments) > 0 {
							for _, attachment := range message.Attachments {
								queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: message.Author}
								queueData.FillMetadata()
								guildData[guild.ID].QueueAdd(queueData)
							}
							responseEmbed = NewGenericEmbed("Clinet Voice", "Added the attached files to the guild queue.")
						} else {
							isPaused, _ := voiceGetPauseState(guild.ID)
							if isPaused {
								_, _ = voiceResume(guild.ID)
								responseEmbed = NewGenericEmbed("Clinet Voice", "Resumed the audio playback.")
							} else {
								responseEmbed = NewErrorEmbed("Clinet Voice Error", "The current audio is already playing.")
							}
						}
					} else {
						if len(message.Attachments) > 0 {
							for _, attachment := range message.Attachments {
								queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: message.Author}
								queueData.FillMetadata()
								guildData[guild.ID].QueueAdd(queueData)
							}
							responseEmbed = NewGenericEmbed("Clinet Voice", "Added the attached files to the guild queue. Use ``"+botData.CommandPrefix+"play`` to begin playback from the beginning of the queue.")
						} else {
							if len(guildData[guild.ID].AudioQueue) > 0 {
								queueData := guildData[guild.ID].AudioQueue[0]
								queueData.FillMetadata()
								guildData[guild.ID].QueueRemove(0)
								go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
							} else {
								responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must specify either a YouTube search query or a YouTube/SoundCloud/direct URL to play.")
							}
						}
					}
				} else if len(cmd) == 2 { //One-word query or URL was specified
					_, err := url.ParseRequestURI(cmd[1]) //Check to see if first parameter is URL
					if err != nil {                       //First parameter is not URL
						queryURL, err := voiceGetQuery(cmd[1])
						if err != nil {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error getting a result for the specified query.")
						} else {
							queueData := AudioQueueEntry{MediaURL: queryURL, Requester: message.Author, Type: "youtube"}
							queueData.FillMetadata()
							if voiceIsStreaming(guild.ID) {
								guildData[guild.ID].QueueAdd(queueData)
								responseEmbed = queueData.GetQueueAddedEmbed()
							} else {
								guildData[guild.ID].AudioNowPlaying = queueData
								responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
								go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
							}
						}
					} else { //First parameter is URL
						queueData := AudioQueueEntry{MediaURL: cmd[1], Requester: message.Author}
						queueData.FillMetadata()
						if voiceIsStreaming(guild.ID) {
							guildData[guild.ID].QueueAdd(queueData)
							responseEmbed = queueData.GetQueueAddedEmbed()
						} else {
							guildData[guild.ID].AudioNowPlaying = queueData
							responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
							go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
						}
					}
				} else if len(cmd) >= 3 { //Multi-word query was specified
					query := strings.Join(cmd[1:], " ") //Get the full search query without the play command
					queryURL, err := voiceGetQuery(query)
					if err != nil {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error getting a result for the specified query.")
					} else {
						queueData := AudioQueueEntry{MediaURL: queryURL, Requester: message.Author, Type: "youtube"}
						queueData.FillMetadata()
						if voiceIsStreaming(guild.ID) {
							guildData[guild.ID].QueueAdd(queueData)
							responseEmbed = queueData.GetQueueAddedEmbed()
						} else {
							guildData[guild.ID].AudioNowPlaying = queueData
							responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
							go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
						}
					}
				}
			} else {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel to use before using the play command.")
			}
			guildData[guild.ID].VoiceData.IsPlaybackPreparing = false
		case "stop":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					if voiceIsStreaming(guild.ID) {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Stopped the audio playback.")
						voiceStop(guild.ID)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					foundVoiceChannel = true
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the stop command.")
			}
		case "skip":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					if voiceIsStreaming(guild.ID) {
						voiceSkip(guild.ID)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					foundVoiceChannel = true
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the skip command.")
			}
		case "pause":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					isPaused, err := voicePause(guild.ID)
					if err != nil {
						if isPaused == false {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
						} else {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "The current audio is already paused.")
						}
					} else {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Paused the audio playback.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the pause command.")
			}
		case "resume":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					isPaused, err := voiceResume(guild.ID)
					if err != nil {
						if isPaused == false {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
						} else {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "The current audio is already playing.")
						}
					} else {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Resumed the audio playback.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the resume command.")
			}
		case "repeat":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					if voiceIsStreaming(guild.ID) {
						switch guildData[guild.ID].VoiceData.RepeatLevel {
						case 0: //No Repeat
							guildData[guild.ID].VoiceData.RepeatLevel = 1
							responseEmbed = NewGenericEmbed("Clinet Voice", "The current guild queue will be repeated.")
						case 1: //Repeat Playlist
							guildData[guild.ID].VoiceData.RepeatLevel = 2
							responseEmbed = NewGenericEmbed("Clinet Voice", "The currently now playing audio will be repeated.")
						case 2: //Repeat Now Playing
							guildData[guild.ID].VoiceData.RepeatLevel = 0
							responseEmbed = NewGenericEmbed("Clinet Voice", "The current guild queue will play through as normal.")
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the repeat command.")
			}
		case "shuffle":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					if voiceIsStreaming(guild.ID) {
						newAudioQueue := make([]AudioQueueEntry, len(guildData[guild.ID].AudioQueue))
						permutation := rand.Perm(len(guildData[guild.ID].AudioQueue))
						for i, v := range permutation {
							newAudioQueue[v] = guildData[guild.ID].AudioQueue[i]
						}
						guildData[guild.ID].AudioQueue = newAudioQueue

						responseEmbed = NewGenericEmbed("Clinet Voice", "The current guild queue has been shuffled.")
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the shuffle command.")
			}
		case "youtube", "yt":
			if len(cmd) > 1 {
				switch cmd[1] {
				case "search", "s":
					if guildData[guild.ID] == nil {
						guildData[guild.ID] = &GuildData{}
						guildData[guild.ID].VoiceData = VoiceData{}
					}
					for guildData[guild.ID].VoiceData.IsPlaybackPreparing {
						//Wait for the handling of a previous playback command to finish
					}
					foundVoiceChannel := false
					for _, voiceState := range guild.VoiceStates {
						if voiceState.UserID == message.Author.ID {
							foundVoiceChannel = true
							voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
							break
						}
					}
					if foundVoiceChannel {
						query := strings.Join(cmd[2:], " ") //Get the full search query without the search command
						if query == "" {
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "You must enter a search query to view the results of.")
						} else {
							_, guildFound := guildData[guild.ID]
							if !guildFound {
								guildData[guild.ID] = &GuildData{}
							}
							if guildData[guild.ID].YouTubeResults == nil {
								guildData[guild.ID].YouTubeResults = make(map[string]*YouTubeResultNav)
							}

							guildData[guild.ID].YouTubeResults[message.Author.ID] = &YouTubeResultNav{}
							page := guildData[guild.ID].YouTubeResults[message.Author.ID]
							err := page.Search(query)
							if err != nil {
								responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "There was an error finding search results for that query.")
							} else {
								commandList := "cli$yt select N - Selects result N"
								if page.PrevPageToken != "" {
									commandList += "\ncli$yt prev - Displays the results for the previous page"
								}
								if page.NextPageToken != "" {
									commandList += "\ncli$yt next - Displays the results for the next page"
								}
								commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

								results, _ := page.GetResults()
								responseEmbed = NewEmbed().
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
							}
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "You must join the voice channel to use before using the YouTube search command.")
					}
				case "next", "n", "+":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						page := guildData[guild.ID].YouTubeResults[message.Author.ID]
						err := page.Next()
						if err != nil {
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "There was an error finding the next page.")
						} else {
							commandList := "cli$yt select N - Selects result N"
							if page.PrevPageToken != "" {
								commandList += "\ncli$yt prev - Displays the results for the previous page"
							}
							if page.NextPageToken != "" {
								commandList += "\ncli$yt next - Displays the results for the next page"
							}
							commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

							results, _ := page.GetResults()
							responseEmbed = NewEmbed().
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
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				case "prev", "previous", "p", "-":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						page := guildData[guild.ID].YouTubeResults[message.Author.ID]
						err := page.Prev()
						if err != nil {
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "There was an error finding the previous page.")
						} else {
							commandList := "cli$yt select N - Selects result N"
							if page.PrevPageToken != "" {
								commandList += "\ncli$yt prev - Displays the results for the previous page"
							}
							if page.NextPageToken != "" {
								commandList += "\ncli$yt next - Displays the results for the next page"
							}
							commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

							results, _ := page.GetResults()
							responseEmbed = NewEmbed().
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
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				case "cancel", "c":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						guildData[guild.ID].YouTubeResults[message.Author.ID] = nil
						responseEmbed = NewGenericEmbed("Clinet YouTube Search", "Cancelled the search.")
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				case "select":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						page := guildData[guild.ID].YouTubeResults[message.Author.ID]
						results, _ := page.GetResults()

						selection, err := strconv.Atoi(cmd[2])
						if err != nil { //Specified selection is not a valid integer
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "``"+cmd[2]+"`` is not a valid number.")
						} else {
							if selection > len(results) || selection <= 0 {
								responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "An invalid selection was specified.")
							} else {
								result := results[selection-1]
								resultURL := "https://youtube.com/watch?v=" + result.Id.VideoId

								queueData := AudioQueueEntry{MediaURL: resultURL, Requester: message.Author, Type: "youtube"}
								queueData.FillMetadata()
								if voiceIsStreaming(guild.ID) {
									guildData[guild.ID].QueueAdd(queueData)
									responseEmbed = queueData.GetQueueAddedEmbed()
								} else {
									guildData[guild.ID].AudioNowPlaying = queueData
									responseEmbed = queueData.GetNowPlayingEmbed()
									go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
								}
							}
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				}
			}
		case "queue":
			if len(cmd) > 1 {
				switch cmd[1] {
				case "help":
					responseEmbed = NewEmbed().
						SetTitle(botData.BotName+" - Queue Help").
						SetDescription("A list of available queue commands for "+botData.BotName+".").
						AddField(botData.CommandPrefix+"queue help", "Displays this help message.").
						AddField(botData.CommandPrefix+"queue clear", "Clears the current queue.").
						AddField(botData.CommandPrefix+"queue list", "Lists all entries in the queue.").
						AddField(botData.CommandPrefix+"queue remove (entry)", "Removes a specified entry from the queue.").
						SetColor(0xFAFAFA).MessageEmbed
				case "clear":
					if guildData[guild.ID].AudioQueue == nil {
						guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
					}
					if len(guildData[guild.ID].AudioQueue) > 0 {
						guildData[guild.ID].QueueClear()
						responseEmbed = NewGenericEmbed("Clinet Queue", "Cleared the queue.")
					} else {
						responseEmbed = NewErrorEmbed("Clinet Queue Error", "There are no entries in the queue to clear.")
					}
				case "list":
					if guildData[guild.ID].AudioQueue == nil {
						guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
					}
					if len(guildData[guild.ID].AudioQueue) > 0 {
						queueList := ""
						for queueEntryNumber, queueEntry := range guildData[guild.ID].AudioQueue {
							displayNumber := strconv.Itoa(queueEntryNumber + 1)
							if queueList != "" {
								queueList += "\n"
							}
							switch queueEntry.Type {
							case "youtube", "soundcloud":
								queueList += displayNumber + ". ``" + queueEntry.Title + "`` by ``" + queueEntry.Author + "``\n\tRequested by " + queueEntry.Requester.String()
							default:
								queueList += displayNumber + ". ``" + queueEntry.MediaURL + "``\n\tRequested by " + queueEntry.Requester.String()
							}
						}
						responseEmbed = NewGenericEmbed("Queue for "+guild.Name, queueList)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Queue Error", "There are no entries in the queue to list.")
					}
				case "remove":
					if guildData[guild.ID].AudioQueue == nil {
						guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
					}
					if len(cmd) > 2 {
						invalidQueueEntry := ""
						for _, queueEntry := range cmd[2:] { //Range over all specified queue entries
							queueEntryNumber, err := strconv.Atoi(queueEntry)
							if err != nil { //Specified queue entry is not a valid integer
								invalidQueueEntry = queueEntry
								break
							} else {
								queueEntryNumber -= 1 //Compensate for 0-index
							}

							if queueEntryNumber > len(guildData[guild.ID].AudioQueue) || queueEntryNumber < 0 {
								invalidQueueEntry = queueEntry
								break
							}
						}
						if invalidQueueEntry != "" {
							responseEmbed = NewErrorEmbed("Clinet Queue Error", invalidQueueEntry+" is not a valid queue entry.")
						} else {
							var newAudioQueue []AudioQueueEntry
							for queueEntryN, queueEntry := range guildData[guild.ID].AudioQueue {
								keepQueueEntry := true
								for _, removedQueueEntry := range cmd[2:] {
									removedQueueEntryNumber, _ := strconv.Atoi(removedQueueEntry)
									removedQueueEntryNumber -= 1
									if queueEntryN == removedQueueEntryNumber {
										keepQueueEntry = false
										break
									}
								}
								if keepQueueEntry {
									newAudioQueue = append(newAudioQueue, queueEntry)
								}
							}

							guildData[guild.ID].AudioQueue = newAudioQueue

							if len(cmd) > 3 {
								responseEmbed = NewGenericEmbed("Clinet Queue", "Successfully removed the specified queue entries.")
							} else {
								responseEmbed = NewGenericEmbed("Clinet Queue", "Successfully removed the specified queue entry.")
							}
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet Queue Error", "You must specify which entries to remove from the queue.")
					}
				}
			} else {
				if guildData[guild.ID].AudioQueue == nil {
					guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
				}
				if len(guildData[guild.ID].AudioQueue) > 0 {
					queueList := ""
					for queueEntryNumber, queueEntry := range guildData[guild.ID].AudioQueue {
						displayNumber := strconv.Itoa(queueEntryNumber + 1)
						if queueList != "" {
							queueList += "\n"
						}
						switch queueEntry.Type {
						case "youtube", "soundcloud":
							queueList += displayNumber + ". ``" + queueEntry.Title + "`` by ``" + queueEntry.Author + "``\n\tRequested by " + queueEntry.Requester.String()
						default:
							queueList += displayNumber + ". ``" + queueEntry.MediaURL + "``\n\tRequested by " + queueEntry.Requester.String()
						}
					}
					responseEmbed = NewGenericEmbed("Queue for "+guild.Name, queueList)
				} else {
					responseEmbed = NewErrorEmbed("Clinet Queue Error", "There are no entries in the queue to list.")
				}
			}
		case "nowplaying", "np":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					if voiceIsStreaming(guild.ID) {
						//Create and display now playing embed
						responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingDurationEmbed(guildData[guild.ID].VoiceData.StreamingSession)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					foundVoiceChannel = true
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the now playing command.")
			}
		}
	} else {
		regexpBotName, _ := regexp.MatchString("<(@|@\\!)"+session.State.User.ID+">(.*?)", content)
		if regexpBotName && strings.HasSuffix(content, "?") {
			if !updatedMessageEvent {
				typingEvent(session, message.ChannelID)
			}
			query := content

			query = strings.Replace(query, "<@!"+session.State.User.ID+">", "", -1)
			query = strings.Replace(query, "<@"+session.State.User.ID+">", "", -1)
			for {
				if strings.HasPrefix(query, " ") {
					query = strings.Replace(query, " ", "", 1)
				} else if strings.HasPrefix(query, ",") {
					query = strings.Replace(query, ",", "", 1)
				} else if strings.HasPrefix(query, ":") {
					query = strings.Replace(query, ":", "", 1)
				} else {
					break
				}
			}

			usedCustomResponse := false
			if len(botData.CustomResponses) > 0 {
				for _, response := range botData.CustomResponses {
					regexpMatched, _ := regexp.MatchString(response.Expression, query)
					if regexpMatched {
						random := rand.Intn(len(response.Responses))
						responseEmbed = NewGenericEmbed("Clinet Response", response.Responses[random].Response)
						usedCustomResponse = true
					}
				}
			}
			if usedCustomResponse == false {
				responseEmbed, err = queryDuckDuckGo(query)
				if err != nil {
					responseEmbed, err = queryWolframAlpha(query)
					if err != nil {
						responseEmbed = NewErrorEmbed("Query Error", "There was an error finding the data you requested.")
					}
				}
			}
		}
	}

	if responseEmbed != nil {
		if !updatedMessageEvent {
			typingEvent(session, message.ChannelID)
		}

		fixedEmbed := Embed{responseEmbed}
		fixedEmbed.InlineAllFields().Truncate()
		responseEmbed = fixedEmbed.MessageEmbed

		canUpdateMessage := false
		responseID := ""

		_, guildFound := guildData[guild.ID]
		if guildFound {
			if guildData[guild.ID].Queries != nil {
				if guildData[guild.ID].Queries[message.ID] != nil {
					debugLog("> Found previous response", false)
					canUpdateMessage = true
					responseID = guildData[guild.ID].Queries[message.ID].ResponseMessageID
				} else {
					debugLog("> Previous response not found, initializing...", false)
					guildData[guild.ID].Queries[message.ID] = &Query{}
				}
			} else {
				debugLog("> Queries not found, initializing...", false)
				guildData[guild.ID].Queries = make(map[string]*Query)
				debugLog("> Previous response not found, initializing...", false)
				guildData[guild.ID].Queries[message.ID] = &Query{}
			}
		} else {
			debugLog("> Guild not found, initializing...", false)
			guildData[guild.ID] = &GuildData{}
			debugLog("> Queries not found, initializing...", false)
			guildData[guild.ID].Queries = make(map[string]*Query)
			debugLog("> Previous response not found, initializing...", false)
			guildData[guild.ID].Queries[message.ID] = &Query{}
		}

		if canUpdateMessage {
			debugLog("> Editing response...", false)
			session.ChannelMessageEditEmbed(message.ChannelID, responseID, responseEmbed)
		} else {
			debugLog("> Sending response...", false)
			responseMessage, err := session.ChannelMessageSendEmbed(message.ChannelID, responseEmbed)
			if err != nil {
				debugLog("> Error sending response message", false)
			} else {
				debugLog("> Storing response...", false)
				guildData[guild.ID].Queries[message.ID].ResponseMessageID = responseMessage.ID
			}
		}
	}
}

func queryImgur(url string) (*discordgo.MessageEmbed, error) {
	imgurInfo, _, err := botData.BotClients.Imgur.GetInfoFromURL(url)
	if err != nil {
		debugLog("[Imgur] Error getting info from URL ["+url+"]", false)
		return nil, errors.New("Error getting info from URL.")
	}
	if imgurInfo.Image != nil {
		debugLog("[Imgur] Detected image from URL ["+url+"]", false)
		imgurImage := imgurInfo.Image
		imgurEmbed := NewEmbed().
			SetTitle(imgurImage.Title).
			SetDescription(imgurImage.Description).
			AddField("Views", strconv.Itoa(imgurImage.Views)).
			AddField("NSFW", strconv.FormatBool(imgurImage.Nsfw)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.Album != nil {
		debugLog("[Imgur] Detected album from URL ["+url+"]", false)
		imgurAlbum := imgurInfo.Album
		imgurEmbed := NewEmbed().
			SetTitle(imgurAlbum.Title).
			SetDescription(imgurAlbum.Description).
			AddField("Uploader", imgurAlbum.AccountURL).
			AddField("Image Count", strconv.Itoa(imgurAlbum.ImagesCount)).
			AddField("Views", strconv.Itoa(imgurAlbum.Views)).
			AddField("NSFW", strconv.FormatBool(imgurAlbum.Nsfw)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.GImage != nil {
		debugLog("[Imgur] Detected gallery image from URL ["+url+"]", false)
		imgurGImage := imgurInfo.GImage
		imgurEmbed := NewEmbed().
			SetTitle(imgurGImage.Title).
			SetDescription(imgurGImage.Description).
			AddField("Topic", imgurGImage.Topic).
			AddField("Uploader", imgurGImage.AccountURL).
			AddField("Views", strconv.Itoa(imgurGImage.Views)).
			AddField("NSFW", strconv.FormatBool(imgurGImage.Nsfw)).
			AddField("Comment Count", strconv.Itoa(imgurGImage.CommentCount)).
			AddField("Upvotes", strconv.Itoa(imgurGImage.Ups)).
			AddField("Downvotes", strconv.Itoa(imgurGImage.Downs)).
			AddField("Points", strconv.Itoa(imgurGImage.Points)).
			AddField("Score", strconv.Itoa(imgurGImage.Score)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.GAlbum != nil {
		debugLog("[Imgur] Detected gallery album from URL ["+url+"]", false)
		imgurGAlbum := imgurInfo.GAlbum
		imgurEmbed := NewEmbed().
			SetTitle(imgurGAlbum.Title).
			SetDescription(imgurGAlbum.Description).
			AddField("Topic", imgurGAlbum.Topic).
			AddField("Uploader", imgurGAlbum.AccountURL).
			AddField("Views", strconv.Itoa(imgurGAlbum.Views)).
			AddField("NSFW", strconv.FormatBool(imgurGAlbum.Nsfw)).
			AddField("Comment Count", strconv.Itoa(imgurGAlbum.CommentCount)).
			AddField("Upvotes", strconv.Itoa(imgurGAlbum.Ups)).
			AddField("Downvotes", strconv.Itoa(imgurGAlbum.Downs)).
			AddField("Points", strconv.Itoa(imgurGAlbum.Points)).
			AddField("Score", strconv.Itoa(imgurGAlbum.Score)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else {
		debugLog("[Imgur] Error detecting Imgur type from URL ["+url+"]", false)
		return nil, errors.New("Error detecting Imgur URL type.")
	}
	return nil, errors.New("Error detecting Imgur URL type.")
}

func queryWolframAlpha(query string) (*discordgo.MessageEmbed, error) {
	debugLog("[Wolfram|Alpha] Getting result for query ["+query+"]...", false)
	queryResultData, err := botData.BotClients.Wolfram.GetQueryResult(query, nil)
	if err != nil {
		debugLog("[Wolfram|Alpha] Error getting query result: "+fmt.Sprintf("%v", err), false)
		return nil, errors.New("Error getting response from Wolfram|Alpha.")
	}

	result := queryResultData.QueryResult
	pods := result.Pods
	if len(pods) == 0 {
		debugLog("[Wolfram|Alpha] Error getting pods from query", false)
		return nil, errors.New("Error getting pods from query.")
	}

	fields := []*discordgo.MessageEmbedField{}

	for _, pod := range pods {
		podTitle := pod.Title
		if wolframIsPodDenied(podTitle) {
			debugLog("[Wolfram|Alpha] Denied pod: "+podTitle, false)
			continue
		}

		subPods := pod.SubPods
		if len(subPods) > 0 { //Skip this pod if no subpods are found
			for _, subPod := range subPods {
				plaintext := subPod.Plaintext
				if plaintext != "" {
					fields = append(fields, &discordgo.MessageEmbedField{Name: podTitle, Value: plaintext})
				}
			}
		}
	}

	if len(fields) == 0 { //No results were found
		debugLog("[Wolfram|Alpha] Error getting legal data from Wolfram|Alpha", false)
		return nil, errors.New("Error getting legal data from Wolfram|Alpha.")
	} else {
		wolframEmbed := NewEmbed().
			SetColor(0xDA0E1A).
			SetFooter("Results from Wolfram|Alpha.", "https://upload.wikimedia.org/wikipedia/en/thumb/8/83/Wolfram_Alpha_December_2016.svg/257px-Wolfram_Alpha_December_2016.svg.png").MessageEmbed
		wolframEmbed.Fields = fields
		return wolframEmbed, nil
	}
}
func wolframIsPodDenied(podTitle string) bool {
	for _, deniedPodTitle := range botData.BotOptions.WolframDeniedPods {
		if deniedPodTitle == podTitle {
			return true //Pod is denied
		}
	}
	return false //Pod is not denied
}

func queryDuckDuckGo(query string) (*discordgo.MessageEmbed, error) {
	debugLog("[DuckDuckGo] Getting result for query ["+query+"]...", false)
	queryResult, err := botData.BotClients.DuckDuckGo.GetQueryResult(query)
	if err != nil {
		debugLog("[DuckDuckGo] Error getting query result: "+fmt.Sprintf("%v", err), false)
		return nil, errors.New("Error getting response from DuckDuckGo.")
	}

	result := ""
	if queryResult.Definition != "" {
		result = queryResult.Definition
	} else if queryResult.Answer != "" {
		result = queryResult.Answer
	} else if queryResult.AbstractText != "" {
		result = queryResult.AbstractText
	}
	if result == "" {
		debugLog("[DuckDuckGo] Error getting allowed result from response", false)
		return nil, errors.New("Error getting allowed result from response")
	}

	duckduckgoEmbed := NewEmbed().
		SetTitle(queryResult.Heading).
		SetDescription(result).
		SetColor(0xDF5730).
		SetFooter("Results from DuckDuckGo.", "https://upload.wikimedia.org/wikipedia/en/9/90/The_DuckDuckGo_Duck.png").MessageEmbed
	if queryResult.Image != "" {
		duckduckgoEmbed.Image = &discordgo.MessageEmbedImage{URL: queryResult.Image}
	}
	return duckduckgoEmbed, nil
}

func voiceJoin(session *discordgo.Session, guildID, channelID, channelIDJoinedFrom string) error {
	_, guildFound := guildData[guildID]
	if guildFound {
		if guildData[guildID].VoiceData.VoiceConnection != nil {
			if guildData[guildID].VoiceData.VoiceConnection.ChannelID == channelID {
				debugLog("> Found previous matching voice connection, staying...", false)
				return nil //We're already in the selected voice channel
			} else {
				debugLog("> Found previous mismatch voice connection, leaving...", false)
				err := voiceLeave(guildID, channelID)
				if err != nil {
					return errors.New("Error leaving specified voice channel")
				}
			}
		}
	} else {
		debugLog("> Guild data not found, initializing...", false)
		guildData[guildID] = &GuildData{}
		guildData[guildID].VoiceData = VoiceData{}
	}
	voiceConnection, err := session.ChannelVoiceJoin(guildID, channelID, false, false)
	if err != nil {
		return errors.New("Error joining specified voice channel.")
	} else {
		guildData[guildID].VoiceData.VoiceConnection = voiceConnection
		guildData[guildID].VoiceData.ChannelIDJoinedFrom = channelIDJoinedFrom
		return nil
	}
}

func voiceLeave(guildID, channelID string) error {
	_, guildFound := guildData[guildID]
	if guildFound {
		if guildData[guildID].VoiceData.VoiceConnection != nil {
			debugLog("> Found previous voice connection, leaving...", false)
			guildData[guildID].VoiceData.VoiceConnection.Disconnect()
			//			guildData[guildID].VoiceData = VoiceData{}
			return nil
		} else {
			return errors.New("Not connected to specified voice channel.")
		}
	} else {
		return errors.New("Not connected to specified voice channel.")
	}
}

func voicePlay(guildID, mediaURL string) error {
	if guildData[guildID].VoiceData.VoiceConnection == nil {
		return errors.New("Not connected to a voice channel.")
	}

	_, err := url.ParseRequestURI(mediaURL)
	if err != nil {
		return errors.New("Specified URL is invalid.")
	}

	mediaURL, err = getMediaURL(mediaURL)
	if err != nil {
		return err
	}

	//Setup pointers to guild data for local usage
	//var voiceConnection *discordgo.VoiceConnection = guildData[guildID].VoiceData.VoiceConnection
	//var encodingSession *dca.EncodeSession = guildData[guildID].VoiceData.EncodingSession
	//var streamingSession *dca.StreamingSession = guildData[guildID].VoiceData.StreamingSession

	//Setup the audio encoding options
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	//Create the encoding session to encode the audio to DCA in a stream
	guildData[guildID].VoiceData.EncodingSession, err = dca.EncodeFile(mediaURL, options)
	if err != nil {
		debugLog("[Voice] Error encoding file ["+mediaURL+"]: "+fmt.Sprintf("%v", err), false)
		return errors.New("Error encoding specified URL to DCA audio.")
	}

	//Set speaking to true
	guildData[guildID].VoiceData.VoiceConnection.Speaking(true)

	//Make a channel for signals when playback is finished
	done := make(chan error)

	//Create the audio stream
	//streamingSession = dca.NewStream(encodingSession, voiceConnection, done)
	guildData[guildID].VoiceData.StreamingSession = dca.NewStream(guildData[guildID].VoiceData.EncodingSession, guildData[guildID].VoiceData.VoiceConnection, done)

	//Set playback running bool to true
	guildData[guildID].VoiceData.IsPlaybackRunning = true

	//Set playback stopped manually bool to false
	guildData[guildID].VoiceData.WasStoppedManually = false

	for guildData[guildID].VoiceData.IsPlaybackRunning {
		select {
		case err := <-done:
			if err != nil {
				guildData[guildID].VoiceData.IsPlaybackRunning = false
			}
		}
	}

	debugLog("-- Playback finished", false)
	debugLog("-- Status: "+strconv.FormatBool(guildData[guildID].VoiceData.IsPlaybackRunning)+"|"+strconv.FormatBool(guildData[guildID].VoiceData.WasStoppedManually)+"|"+strconv.FormatBool(guildData[guildID].VoiceData.WasSkipped), false)

	//Set speaking to false
	guildData[guildID].VoiceData.VoiceConnection.Speaking(false)

	//Check streaming session for why playback stopped
	_, err = guildData[guildID].VoiceData.StreamingSession.Finished()

	//Clean up streaming session
	guildData[guildID].VoiceData.StreamingSession = nil

	//Clean up encoding session
	guildData[guildID].VoiceData.EncodingSession.Stop()
	guildData[guildID].VoiceData.EncodingSession.Cleanup()
	guildData[guildID].VoiceData.EncodingSession = nil

	//If playback stopped from an error, return that error
	if err != nil {
		debugLog("-- Playback error", false)
		return err
	}
	return nil
}

func voicePlayWrapper(session *discordgo.Session, guildID, channelID, mediaURL string) {

	//0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing

	err := voicePlay(guildID, mediaURL)
	if guildData[guildID].VoiceData.RepeatLevel == 2 { //Repeat Now Playing
		for guildData[guildID].VoiceData.RepeatLevel == 2 {
			err = voicePlay(guildID, mediaURL)
			if err != nil {
				guildData[guildID].AudioNowPlaying = AudioQueueEntry{} //Clear now playing slot
				errorEmbed := NewErrorEmbed("Clinet Voice Error", "There was an error playing the specified audio.")
				session.ChannelMessageSendEmbed(channelID, errorEmbed)
				return
			}
		}
	}
	if guildData[guildID].VoiceData.RepeatLevel == 1 { //Repeat Playlist
		guildData[guildID].QueueAdd(guildData[guildID].AudioNowPlaying) //Shift the now playing entry to the end of the guild queue
	}
	guildData[guildID].AudioNowPlaying = AudioQueueEntry{} //Clear now playing slot
	if err != nil {
		errorEmbed := NewErrorEmbed("Clinet Voice Error", "There was an error playing the specified audio.")
		session.ChannelMessageSendEmbed(channelID, errorEmbed)
		return
	} else {
		if guildData[guildID].VoiceData.WasStoppedManually {
			guildData[guildID].VoiceData.WasStoppedManually = false
		} else if guildData[guildID].VoiceData.IsPlaybackRunning == false || guildData[guildID].VoiceData.WasSkipped == true {
			guildData[guildID].VoiceData.WasSkipped = false //Reset skip bool in case it was true

			//When the song finishes playing, we should run on a loop to make sure the next songs continue playing
			for len(guildData[guildID].AudioQueue) > 0 {
				//Move next guild queue entry into now playing slot
				guildData[guildID].AudioNowPlaying = guildData[guildID].AudioQueue[0]
				guildData[guildID].QueueRemove(0)

				//Create and display now playing embed
				nowPlayingEmbed := guildData[guildID].AudioNowPlaying.GetNowPlayingEmbed()
				session.ChannelMessageSendEmbed(channelID, nowPlayingEmbed)

				//Play audio
				err := voicePlay(guildID, guildData[guildID].AudioNowPlaying.MediaURL)
				if guildData[guildID].VoiceData.RepeatLevel == 2 { //Repeat Now Playing
					for guildData[guildID].VoiceData.RepeatLevel == 2 {
						err = voicePlay(guildID, guildData[guildID].AudioNowPlaying.MediaURL)
						if err != nil {
							guildData[guildID].AudioNowPlaying = AudioQueueEntry{} //Clear now playing slot
							errorEmbed := NewErrorEmbed("Clinet Voice Error", "There was an error playing the specified audio.")
							session.ChannelMessageSendEmbed(channelID, errorEmbed)
							return
						}
					}
				}
				if guildData[guildID].VoiceData.RepeatLevel == 1 { //Repeat Playlist
					guildData[guildID].QueueAdd(guildData[guildID].AudioNowPlaying) //Shift the now playing entry to the end of the guild queue
				}
				guildData[guildID].AudioNowPlaying = AudioQueueEntry{} //Clear now playing slot
				if err != nil {
					errorEmbed := NewErrorEmbed("Clinet Voice Error", "There was an error playing the specified audio.")
					session.ChannelMessageSendEmbed(channelID, errorEmbed)
					return //Prevent next guild queue entry from playing
				} else {
					if guildData[guildID].VoiceData.WasStoppedManually {
						guildData[guildID].VoiceData.WasStoppedManually = false
						return //Prevent next guild queue entry from playing
					}
				}
			}
		}
	}
}

func voiceStop(guildID string) {
	if guildData[guildID] != nil {
		guildData[guildID].VoiceData.WasStoppedManually = true //Make sure other threads know it was stopped manually
		guildData[guildID].VoiceData.EncodingSession.Stop()    //Stop the encoding session manually
		guildData[guildID].VoiceData.IsPlaybackRunning = false //Let the voice play function clean up on its own
	}
}

func voiceSkip(guildID string) {
	if guildData[guildID] != nil {
		guildData[guildID].VoiceData.WasSkipped = true      //Let the voice play wrapper function continue to the next song if available
		guildData[guildID].VoiceData.EncodingSession.Stop() //Stop the encoding session manually
	}
}

func voiceIsStreaming(guildID string) bool {
	if guildData[guildID] == nil {
		return false
	}
	return guildData[guildID].VoiceData.IsPlaybackRunning
}

func voiceGetPauseState(guildID string) (bool, error) {
	if guildData[guildID].VoiceData.StreamingSession == nil {
		return false, errors.New("Could not find the streaming session for the specified guild.")
	}

	isPaused := guildData[guildID].VoiceData.StreamingSession.Paused()
	return isPaused, nil
}

func voicePause(guildID string) (bool, error) {
	if guildData[guildID].VoiceData.StreamingSession == nil {
		return false, errors.New("Could not find the streaming session for the specified guild.")
	}

	isPaused := guildData[guildID].VoiceData.StreamingSession.Paused()
	if isPaused {
		return true, errors.New("The specified guild's streaming session is already paused.")
	}

	guildData[guildID].VoiceData.StreamingSession.SetPaused(true)
	return true, nil
}

func voiceResume(guildID string) (bool, error) {
	if guildData[guildID].VoiceData.StreamingSession == nil {
		return false, errors.New("Could not find the streaming session for the specified guild.")
	}

	isPaused := guildData[guildID].VoiceData.StreamingSession.Paused()
	if isPaused {
		guildData[guildID].VoiceData.StreamingSession.SetPaused(false)
		return true, nil
	}

	return true, errors.New("The specified guild's streaming session is already playing.")
}

func voiceGetQuery(query string) (string, error) {
	call := botData.BotClients.YouTube.Search.List("id").
		Q(query).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return "", errors.New("Could not find any results for the specified query.")
	}

	for _, item := range response.Items {
		if item.Id.Kind == "youtube#video" {
			url := "https://youtube.com/watch?v=" + item.Id.VideoId
			return url, nil
		}
	}

	return "", errors.New("Could not find a video result for the specified query.")
}

func getMediaURL(url string) (string, error) {
	if isYouTubeURL(url) {
		videoInfo, err := ytdl.GetVideoInfo(url)
		if err != nil {
			return url, err
		}

		format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]

		mediaURL, err := videoInfo.GetDownloadURL(format)
		if err != nil {
			return url, err
		}

		return mediaURL.String(), nil
	}

	if isSoundCloudURL(url) {
		audioInfo, err := botData.BotClients.SoundCloud.GetTrackInfo(url)
		if err != nil {
			return url, err
		}

		return audioInfo.DownloadURL, nil
	}

	return url, nil
}

func isYouTubeURL(url string) bool {
	regexpHasYouTube, _ := regexp.MatchString("(?:https?:\\/\\/)?(?:www\\.)?youtu\\.?be(?:\\.com)?\\/?.*(?:watch|embed)?(?:.*v=|v\\/|\\/)(?:[\\w-_]+)", url)
	if regexpHasYouTube {
		return true
	}
	return false
}
func isSoundCloudURL(url string) bool {
	regexpHasSoundCloud, _ := regexp.MatchString("^(https?:\\/\\/)?(www.)?(m\\.)?soundcloud\\.com\\/[\\w\\-\\.]+(\\/)+[\\w\\-\\.]+/?$", url)
	if regexpHasSoundCloud {
		return true
	}
	return false
}

func updateRandomStatus(session *discordgo.Session, status int) {
	if status == 0 {
		status = rand.Intn(len(botData.CustomStatuses)) + 1
	}
	status -= 1

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

type CaseInsensitiveReplacer struct {
	toReplace   *regexp.Regexp
	replaceWith string
}

func NewCaseInsensitiveReplacer(toReplace, with string) *CaseInsensitiveReplacer {
	return &CaseInsensitiveReplacer{
		toReplace:   regexp.MustCompile("(?i)" + toReplace),
		replaceWith: with,
	}
}
func (cir *CaseInsensitiveReplacer) Replace(str string) string {
	return cir.toReplace.ReplaceAllString(str, cir.replaceWith)
}

func GitHubFetchUser(username string) (*github.User, error) {
	user, _, err := botData.BotClients.GitHub.Users.Get(context.Background(), username)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func GitHubFetchRepo(owner string, repository string) (*github.Repository, error) {
	repo, _, err := botData.BotClients.GitHub.Repositories.Get(context.Background(), owner, repository)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func zeroPad(str string) (result string) {
	if len(str) < 2 {
		result = "0" + str
	} else {
		result = str
	}
	return
}

func secondsToHuman(input float64) (result string) {
	hours := math.Floor(float64(input) / 60 / 60)
	seconds := int(input) % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = int(input) % 60

	if hours > 0 {
		result = strconv.Itoa(int(hours)) + ":" + zeroPad(strconv.Itoa(int(minutes))) + ":" + zeroPad(strconv.Itoa(int(seconds)))
	} else {
		result = zeroPad(strconv.Itoa(int(minutes))) + ":" + zeroPad(strconv.Itoa(int(seconds)))
	}

	return
}

func roundTime(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
