package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/paked/configure" //Allows configuration of the program via external sources
	"github.com/bwmarrin/discordgo" //Allows usage of the Discord API
	"github.com/JoshuaDoes/go-wolfram" //Allows usage of the Wolfram|Alpha API
	"github.com/jonas747/dca" //Allows the encoding/decoding of the Discord Audio format
	"github.com/rylio/ytdl" //Allows the fetching of YouTube video metadata and download URLs
	"google.golang.org/api/googleapi/transport" //Allows the making of authenticated API requests to Google
	"google.golang.org/api/youtube/v3" //Allows usage of the YouTube API
	"github.com/nishanths/go-xkcd" //Allows the fetching of XKCD comics
	"github.com/JoshuaDoes/duckduckgolang" //Allows the usage of the DuckDuckGo API
	"github.com/koffeinsource/go-imgur" //Allows usage of the Imgur API
	"github.com/koffeinsource/go-klogger" //For some reason, this is required for go-imgur's logging
	"github.com/robfig/cron" //Allows for better management of running tasks at specific intervals
	"github.com/JoshuaDoes/go-soundcloud" //Allows usage of the SoundCloud API
)
var ( //Used during development, delete later when all imports are in use
	_ = rand.Intn
	_ = url.ParseRequestURI
	_ = strconv.Itoa
	_ = strings.Replace
	_ = time.NewTicker
	_ = ytdl.GetVideoInfo
	_ = &transport.APIKey{}
	_ = youtube.New
	_ = xkcd.NewClient
	_ = cron.New
)

//Bot data structs
type BotClients struct {
	DuckDuckGo *duckduckgo.Client
	Imgur imgur.Client
	SoundCloud *soundcloud.Client
	Wolfram *wolfram.Client
	XKCD *xkcd.Client
}
type BotData struct {
	BotClients BotClients
	BotKeys BotKeys `json:"botKeys"`
	BotName string `json:"botName"`
	BotOptions BotOptions `json:"botOptions"`
	BotToken string `json:"botToken"`
	CommandPrefix string `json:"cmdPrefix"`
	CustomResponses []CustomResponse `json:"customResponses"`
	DebugMode bool `json:"debugMode"`
}
type BotKeys struct {
	DuckDuckGoAppName string `json:"ddgAppName"`
	ImgurClientID string `json:"imgurClientID"`
	SoundCloudAppVersion string `json:"soundcloudAppVersion"`
	SoundCloudClientID string `json:"soundcloudClientID"`
	WolframAppID string `json:"wolframAppID"`
	YouTubeAPIKey string `json:"youtubeAPIKey"`
}
type BotOptions struct {
	SendTypingEvent bool `json:"sendTypingEvent"`
	UseDuckDuckGo bool `json:"useDuckDuckGo"`
	UseImgur bool `json:"useImgur"`
	UseSoundCloud bool `json:"useSoundCloud"`
	UseWolframAlpha bool `json:"useWolframAlpha"`
	UseXKCD bool `json:"useXKCD"`
	UseYouTube bool `json:"useYouTube"`
	WolframDeniedPods []string `json:"wolframDeniedPods"`
}
type CustomResponse struct {
	Expression string `json:"expression"`
	Regexp *regexp.Regexp
	Responses []string `json:"response"`
}
func (configData *BotData) PrepConfig() error {
	// Bot config checks
	if configData.BotName == "" {
		return errors.New("config:{botName: \"\"}")
	}
	if configData.BotToken == "" {
		return errors.New("config:{botName: \"\"}")
	}
	if configData.CommandPrefix == "" {
		return errors.New("config:{cmdPrefix: \"\"}")
	}

	// Bot key checks
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

	// Custom response checks
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
	AudioQueue []AudioQueueEntry
	VoiceData VoiceData
	Queries map[string]*Query //*GuildData.Queries["messageID"] = *Query
}
func (guild *GuildData) QueueAdd(author, imageURL, title, thumbnailURL, mediaURL string, requester *discordgo.User) {
	var queueData AudioQueueEntry
	queueData.Author = author
	queueData.ImageURL = imageURL
	queueData.MediaURL = mediaURL
	queueData.Requester = requester
	queueData.ThumbnailURL = thumbnailURL
	queueData.Title = title
	guild.AudioQueue = append(guild.AudioQueue, queueData)
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
type AudioQueueEntry struct {
	Author string
	ImageURL string
	MediaURL string
	Requester *discordgo.User
	ThumbnailURL string
	Title string
}
type Query struct {
	ResponseMessageID string
}
type VoiceData struct {
	VoiceConnection *discordgo.VoiceConnection
	EncodingSession *dca.EncodeSession
	Stream *dca.StreamingSession

	IsPlaybackRunning bool //Whether or not playback is currently running
	WasStoppedManually bool //Whether or not playback was stopped manually or automatically
}

var (
	botData *BotData = &BotData{}
	guildData = make(map[string] *GuildData)

	conf = configure.New()
	confConfigFile = conf.String("config", "config.json", "The location of the JSON-structured configuration file")
	configFile string = ""
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

	debugLog("> Disconnecting from Discord...", true)
	for _, guildDataRow := range guildData {
		if guildDataRow.VoiceData.VoiceConnection != nil {
			debugLog("> Closing connection to voice channel " + guildDataRow.VoiceData.VoiceConnection.ChannelID + "...", false)
			guildDataRow.VoiceData.VoiceConnection.Close()
		}
	}
	discord.Close()
}

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: " + fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}
	content, err := message.ContentWithMoreMentionsReplaced(session)
	if err != nil {
		debugLog("> Error finding message content", false)
		return //There was an uhoh somewhere
	}

	if message.Author.Bot {
		debugLog("[S] " + message.Author.Username + "#" + message.Author.Discriminator + " [BOT]: " + content, false)
	} else {
		debugLog("[S] " + message.Author.Username + "#" + message.Author.Discriminator + ": " + content, false)
	}

	go handleMessage(session, message)
}
func discordMessageDelete(session *discordgo.Session, event *discordgo.MessageDelete) {
	message := event //Make it easier to keep track of what's happening

	debugLog("[D] ID: " + message.ID, false)

	guildChannel, err := session.Channel(message.ChannelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			_, messageFound := guildData[guildID].Queries[message.ID]
			if messageFound {
				debugLog("> Deleting message...", false)
				session.ChannelMessageDelete(message.ChannelID, guildData[guildID].Queries[message.ID].ResponseMessageID) //Delete the query response message
				guildData[guildID].Queries[message.ID] = nil //Remove the message from the query list
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

	guildChannel, err := session.State.Channel(channelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			for i := 0; i > len(messages); i++ {
				if guildData[guildID].Queries[messages[i]] != nil {
					message, err := session.State.Message(channelID, messages[i])
					if err == nil {
						session.ChannelMessageDelete(message.ChannelID, guildData[guildID].Queries[messages[i]].ResponseMessageID) //Delete the query response message
						guildData[guildID].Queries[messages[i]] = nil //Remove the message from the query list
					}
				}
			}
		}
	}
}
func discordMessageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		return //The bot should never reply to itself
	}
	content, err := message.ContentWithMoreMentionsReplaced(session)
	if err != nil {
		debugLog("> Error finding message content", false)
		return //There was an uhoh somewhere
	}

	if message.Author.Bot {
		debugLog("[U] " + message.Author.Username + "#" + message.Author.Discriminator + " [BOT]: " + content, false)
	} else {
		debugLog("[U] " + message.Author.Username + "#" + message.Author.Discriminator + ": " + content, false)
	}

	go handleMessage(session, message)
}
func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	updateRandomStatus(session, event, 0)
	cronjob := cron.New()
	cronjob.AddFunc("@every 1m", func() { updateRandomStatus(session, event, 0) })
	cronjob.Start()
}

func handleMessage(session *discordgo.Session, message *discordgo.Message) {
	content, err := message.ContentWithMoreMentionsReplaced(session)
	if err != nil {
		debugLog("> Error finding message content", false)
		return //There was an uhoh somewhere
	}
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

	var responseEmbed *discordgo.MessageEmbed

	if strings.HasPrefix(content, botData.CommandPrefix) {
		cmdMsg := strings.Replace(content, botData.CommandPrefix, "", 1)
		cmd := strings.Split(cmdMsg, " ")
		switch cmd[0] {
			case "help":
				responseEmbed = NewEmbed().
					SetTitle(botData.BotName + " - Help").
					SetDescription("A list of available commands for " + botData.BotName + ".").
					AddField(botData.CommandPrefix + "help", "Displays this help message.").
					AddField(botData.CommandPrefix + "about", "Displays information about " + botData.BotName + " and how to use it.").
					AddField(botData.CommandPrefix + "roll", "Rolls a dice.").
					AddField(botData.CommandPrefix + "doubleroll", "Rolls two die.").
					AddField(botData.CommandPrefix + "coinflip", "Flips a coin.").
					AddField(botData.CommandPrefix + "xkcd (comic number|random|latest)", "Displays an xkcd comic depending on the requested type or comic number.").
					AddField(botData.CommandPrefix + "imgur (url)", "Displays info about the specified Imgur image, album, gallery image, or gallery album.").
					AddField(botData.CommandPrefix + "play (url/YouTube search query)", "Plays either the first result from the specified YouTube search query or the specified YouTube/direct audio URL in the user's current voice channel.").
					AddField(botData.CommandPrefix + "stop", "Stops the currently playing audio.").
					AddField(botData.CommandPrefix + "skip", "Stops the currently playing audio, and, if available, attempts to play the next audio in the queue.").
					AddField(botData.CommandPrefix + "queue", "Lists all entries in the queue.").
					AddField(botData.CommandPrefix + "clear", "Clears the current queue.").
					AddField(botData.CommandPrefix + "leave", "Leaves the current voice channel.").
					SetColor(0xfafafa).MessageEmbed
			case "about":
				responseEmbed = NewEmbed().
					SetTitle(botData.BotName + " - About").
					SetDescription(botData.BotName + " is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
					AddField("How can I use " + botData.BotName + " in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
					AddField("How can I help keep " + botData.BotName + " running?", "The best ways to help keep " + botData.BotName + " running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
					AddField("How can I use " + botData.BotName + "?", "There are many ways to make use of " + botData.BotName + ".\n1) Type ``cli$help`` and try using some of the available commands.\n2) Ask " + botData.BotName + " a question, ex: ``" + botData.BotName + ", what time is it?`` or ``" + botData.BotName + ", what is DiscordApp?``.").
					AddField("Invite Link", "https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot").
					AddField("Donation Link", "https://www.paypal.me/JoshuaDoes").
					AddField("Source Code Link", "https://github.com/JoshuaDoes/clinet-discord/").
					SetColor(0x1c1c1c).MessageEmbed
			case "roll":
				random := rand.Intn(6) + 1
				responseEmbed = NewGenericEmbed("Roll", "You rolled a " + strconv.Itoa(random) + "!")
			case "doubleroll", "rolldouble":
				random1 := rand.Intn(6) + 1
				random2 := rand.Intn(6) + 1
				randomTotal := random1 + random2
				responseEmbed = NewGenericEmbed("Double Roll", "You rolled a " + strconv.Itoa(random1) + " and a " + strconv.Itoa(random2) + ". The total is " + strconv.Itoa(randomTotal) + "!")
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
									SetColor(0x96a8c8).MessageEmbed
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
									SetColor(0x96a8c8).MessageEmbed
							}
						default: //Get specified XKCD comic
							comicNumber, err := strconv.Atoi(cmd[1])
							if err != nil { //Specified comic is not a valid integer
								responseEmbed = NewErrorEmbed("XKCD Error", "``" + cmd[1] + "`` is not a valid number.")
							} else {
								comic, err := botData.BotClients.XKCD.Get(comicNumber)
								if err != nil {
									responseEmbed = NewErrorEmbed("XKCD Error", "Could not find XKCD comic #" + cmd[1] + ".")
								} else {
									responseEmbed = NewEmbed().
										SetTitle("xkcd - #" + cmd[1]).
										SetDescription(comic.Title).
										SetImage(comic.ImageURL).
										SetColor(0x96a8c8).MessageEmbed
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
							SetColor(0x96a8c8).MessageEmbed
					}
				}
			default: //Invalid command specified
				responseEmbed = NewErrorEmbed(botData.BotName + " Error", "Unknown command. Type ``cli$help`` for a list of commands.")
		}
	} else {
		regexpBotName, _ := regexp.MatchString("(?i)" + botData.BotName + "(.*?)", content)
		if regexpBotName && strings.HasSuffix(content, "?") {
			typingEvent(session, message.ChannelID)

			query := content

			replace := NewCaseInsensitiveReplacer("Clinet", "")
			query = replace.Replace(query)
			for {
				if strings.HasPrefix(query, " ") {
					query = strings.Replace(query, " ", "", 1)
				} else if strings.HasPrefix(query, ",") {
					query = strings.Replace(query, ",", "", 1)
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
						responseEmbed = NewGenericEmbed("Clinet Response", response.Responses[random])
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
		canUpdateMessage := false
		responseID := ""

		_, guildFound := guildData[guild.ID]
		if guildFound {
			if guildData[guild.ID].Queries[message.ID] != nil {
				debugLog("> Found previous response", false)
				canUpdateMessage = true
				responseID = guildData[guild.ID].Queries[message.ID].ResponseMessageID
			} else {
				debugLog("> Previous response not found, initializing...", false)
				guildData[guild.ID].Queries[message.ID] = &Query{}
			}
		} else {
			debugLog("> Guild not found, initializing...", false)
			guildData[guild.ID] = &GuildData{}
			guildData[guild.ID].Queries = make(map[string] *Query)
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
		debugLog("[Imgur] Error getting info from URL [" + url + "]", false)
		return nil, errors.New("Error getting info from URL.")
	}
	if imgurInfo.Image != nil {
		debugLog("[Imgur] Detected image from URL [" + url + "]", false)
		imgurImage := imgurInfo.Image
		imgurEmbed := NewEmbed().
			SetTitle(imgurImage.Title).
			SetDescription(imgurImage.Description).
			AddField("Views", strconv.Itoa(imgurImage.Views)).
			AddField("NSFW", strconv.FormatBool(imgurImage.Nsfw)).
			SetColor(0x89c623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.Album != nil {
		debugLog("[Imgur] Detected album from URL [" + url + "]", false)
		imgurAlbum := imgurInfo.Album
		imgurEmbed := NewEmbed().
			SetTitle(imgurAlbum.Title).
			SetDescription(imgurAlbum.Description).
			AddField("Uploader", imgurAlbum.AccountURL).
			AddField("Image Count", strconv.Itoa(imgurAlbum.ImagesCount)).
			AddField("Views", strconv.Itoa(imgurAlbum.Views)).
			AddField("NSFW", strconv.FormatBool(imgurAlbum.Nsfw)).
			SetColor(0x89c623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.GImage != nil {
		debugLog("[Imgur] Detected gallery image from URL [" + url + "]", false)
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
			SetColor(0x89c623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.GAlbum != nil {
		debugLog("[Imgur] Detected gallery album from URL [" + url + "]", false)
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
			SetColor(0x89c623).MessageEmbed
		return imgurEmbed, nil
	} else {
		debugLog("[Imgur] Error detecting Imgur type from URL [" + url + "]", false)
		return nil, errors.New("Error detecting Imgur URL type.")
	}
	return nil, errors.New("Error detecting Imgur URL type.")
}

func queryWolframAlpha(query string) (*discordgo.MessageEmbed, error) {
	debugLog("[Wolfram|Alpha] Getting result for query [" + query + "]...", false)
	queryResultData, err := botData.BotClients.Wolfram.GetQueryResult(query, nil)
	if err != nil {
		debugLog("[Wolfram|Alpha] Error getting query result: " + fmt.Sprintf("%v", err), false)
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
			debugLog("[Wolfram|Alpha] Denied pod: " + podTitle, false)
			continue
		}

		subPods := pod.SubPods
		if len(subPods) > 0 { //Skip this pod if no subpods are found
			for _, subPod := range subPods {
				plaintext := subPod.Plaintext
				if plaintext != "" {
					fields = append(fields, &discordgo.MessageEmbedField{Name:podTitle, Value:plaintext})
				}
			}
		}
	}

	if len(fields) == 0 { //No results were found
		debugLog("[Wolfram|Alpha] Error getting legal data from Wolfram|Alpha", false)
		return nil, errors.New("Error getting legal data from Wolfram|Alpha.")
	} else {
		wolframEmbed := NewEmbed().
			SetColor(0xda0e1a).MessageEmbed
		wolframEmbed.Fields = fields
		return wolframEmbed, nil
	}
}
func wolframIsPodDenied(podTitle string) (bool) {
	for _, deniedPodTitle := range botData.BotOptions.WolframDeniedPods {
		if deniedPodTitle == podTitle {
			return true //Pod is denied
		}
	}
	return false //Pod is not denied
}

func queryDuckDuckGo(query string) (*discordgo.MessageEmbed, error) {
	debugLog("[DuckDuckGo] Getting result for query [" + query + "]...", false)
	queryResult, err := botData.BotClients.DuckDuckGo.GetQueryResult(query)
	if err != nil {
		debugLog("[DuckDuckGo] Error getting query result: " + fmt.Sprintf("%v", err), false)
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
		SetColor(0xdf5730).MessageEmbed
	if queryResult.Image != "" {
		duckduckgoEmbed.Image = &discordgo.MessageEmbedImage{URL:queryResult.Image}
	}
	return duckduckgoEmbed, nil
}

func updateRandomStatus(session *discordgo.Session, event *discordgo.Ready, statusType int) {
	/*
	guildCount := len(event.Guilds)
	userCount := 0
	roleCount := 0
	emojiCount := 0
	channelCount := 0
	presenceCount := 0
	for _, guild := range event.Guilds {
		userCount += len(guild.Members)
		roleCount += len(guild.Roles)
		emojiCount += len(guild.Emojis)
		channelCount += len(guild.Channels)
		presenceCount += len(guild.Presences)
	}
	if statusType == 0 { statusType = rand.Intn(6) + 1 }
	switch statusType {
		case 1:
			session.UpdateStatus(0, "in " + strconv.Itoa(guildCount) + " guilds!") //Playing in x guilds!
		case 2:
			session.UpdateListeningStatus(strconv.Itoa(userCount) + " users!") //Listening to x users!
		case 3:
			session.UpdateStatus(0, "with " + strconv.Itoa(roleCount) + " roles!") //Playing with x roles!
		case 4:
			session.UpdateListeningStatus(strconv.Itoa(emojiCount) + " emojis!") //Listening to x emojis!
		case 5:
			session.UpdateListeningStatus(strconv.Itoa(channelCount) + " channels!") //Listening to x channels!
		case 6:
			session.UpdateStatus(0, "with " + strconv.Itoa(presenceCount) + " presences!") //Playing with x presences!
	}
	*/
	session.UpdateStatus(0, "experimentally!")
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