package main

import (
	"fmt"
	//"io"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"
	"strconv"
	"net/http"
	"regexp"
	"math/rand"
	"errors"

	"github.com/paked/configure" // Allows configuration of the program via external sources
	"github.com/bwmarrin/discordgo" // Allows usage of the Discord API
	"github.com/JoshuaDoes/go-wolfram" // Allows usage of the Wolfram|Alpha API
	"github.com/jonas747/dca" // Allows the encoding/decoding of the Discord Audio format
	"github.com/rylio/ytdl" // Allows the fetching of YouTube video metadata and download URLs
	"google.golang.org/api/googleapi/transport" // Allows the making of authenticated API requests to Google
	"google.golang.org/api/youtube/v3" // Allows usage of the YouTube API
	"github.com/nishanths/go-xkcd" // Allows the fetching of XKCD comics
	"github.com/JoshuaDoes/duckduckgolang" // Allows the usage of the DuckDuckGo API
	"github.com/koffeinsource/go-imgur" // Allows usage of the Imgur API
	"github.com/koffeinsource/go-klogger" // For some reason, this is required for go-imgur's logging
)

type message struct {
    ID  string
    ChannelID string   
}
type GuildQueue struct {
	Queue []Queue
}
type Queue struct {
	Name string
	Author string
	Duration int
	ImageURL string
	ThumbnailURL string
	Requester string
	URL string
}
type VoiceData struct {
	VoiceConnection *discordgo.VoiceConnection
	EncodingSession *dca.EncodeSession
	Stream *dca.StreamingSession
	IsPlaybackRunning bool
	WasPlaybackStoppedManually bool
}

var (
	conf = configure.New()
	confBotToken = conf.String("botToken", "", "Bot Token")
	confBotName = conf.String("botName", "", "Bot Name")
	confBotPrefix = conf.String("botPrefix", "", "Bot Prefix")
	confWolframAppID = conf.String("wolframAppID", "", "Wolfram App ID")
	confDuckDuckGoAppName = conf.String("ddgAppName", "", "DuckDuckGo App Name")
	confYouTubeAPIKey = conf.String("youtubeAPIKey", "", "YouTube API Key")
	confImgurClientID = conf.String("imgurClientID", "", "Imgur Client ID")
	confDebugMode = conf.Bool("debugMode", false, "Debug Mode")
	botToken string = ""
	botName string = ""
	botPrefix string = ""
	wolframAppID string = ""
	ddgAppName string = ""
	youtubeAPIKey string = ""
	imgurClientID string = ""
	debugMode bool = false
	
	wolframClient *wolfram.Client
	ddgClient *duckduckgo.Client
	imgurClient imgur.Client
	
	guildCount int
	guilds = make(map[string] string)
	
	/*
	voiceConnections []*discordgo.VoiceConnection
	encodingSessions []*dca.EncodeSession
	streams []*dca.StreamingSession
	playbackRunning []bool
	playbackStopped []bool
	queue = make(map[string] []string)
	*/
	
	queue = make(map[string] *GuildQueue)
	voiceData = make(map[string] *VoiceData)
	
	messages = make(map[string]chan message)
	responses = make(map[string] string)
)

func debugLog(msg string) {
	if debugMode {
		fmt.Println(msg)
	}
}

func init() {
	conf.Use(configure.NewFlag())
	conf.Use(configure.NewJSONFromFile("config.json"))
}

func main() {
	fmt.Println("Clinet-Discord © JoshuaDoes: 2018.\n")

	fmt.Println("> Loading configuration...")
	conf.Parse()
	botToken = *confBotToken
	botName = *confBotName
	botPrefix = *confBotPrefix
	wolframAppID = *confWolframAppID
	ddgAppName = *confDuckDuckGoAppName
	youtubeAPIKey = *confYouTubeAPIKey
	imgurClientID = *confImgurClientID
	debugMode = *confDebugMode
	if (botToken == "" || botName == "" || botPrefix == "" || wolframAppID == "" || ddgAppName == "" || youtubeAPIKey == "" || imgurClientID == "") {
		fmt.Println("> Configuration not properly setup, exiting...")
		return
	} else {
		fmt.Println("> Successfully loaded configuration.")
		debugLog("botToken: " + botToken)
		debugLog("botName: " + botName)
		debugLog("botPrefix: " + botPrefix)
		debugLog("wolframAppID: " + wolframAppID)
		debugLog("ddgAppName: " + ddgAppName)
		debugLog("youtubeAPIKey: " + youtubeAPIKey)
		debugLog("imgurClientID: " + imgurClientID)
		debugLog("debugMode: " + fmt.Sprintf("%t", debugMode))
	}
	
	debugLog("> Creating a new Discord session...")
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session: " + fmt.Sprintf("%v", err))
		return
	}
	
	debugLog("> Registering Ready callback handler...")
	dg.AddHandler(ready)
	debugLog("> Registering MessageCreate callback handler...")
	dg.AddHandler(messageCreate)
	debugLog("> Registering MessageUpdate callback handler...")
	dg.AddHandler(messageUpdate)
	debugLog("> Registering GuildJoin callback handler...")
	dg.AddHandler(guildCreate)
	debugLog("> Registering GuildDelete callback handler...")
	dg.AddHandler(guildDelete)

	fmt.Println("> Establishing a connection to Discord...")
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection: " + fmt.Sprintf("%v", err))
		return
	}
	
	fmt.Println("> Initializing Wolfram...")
	wolframClient = &wolfram.Client{AppID:wolframAppID}
	
	fmt.Println("> Initializing DuckDuckGo...")
	ddgClient = &duckduckgo.Client{AppName:ddgAppName}
	
	fmt.Println("> Initializing Imgur...")
	imgurClient.HTTPClient = &http.Client{}
	imgurClient.Log = &klogger.CLILogger{}
	imgurClient.ImgurClientID = imgurClientID

	fmt.Println("> " + botName + " has started successfully.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	
	debugLog("> Closing any active voice connections...")
	for _, voiceDataRow := range voiceData {
		voiceDataRow.VoiceConnection.Close()
	}
	
	fmt.Println("> Closing Discord session...")
	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	guildCount = len(s.State.Guilds)
	debugLog("Server count: " + strconv.Itoa(guildCount))
	s.UpdateStatus(0, "in " + strconv.Itoa(guildCount) + " servers!")

	guilds = make(map[string] string)

	for _, guildRow := range s.State.Guilds {
		guilds[guildRow.ID] = guildRow.Name
		debugLog(guildRow.ID + ": " + guildRow.Name)
	}
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildCount = len(s.State.Guilds)
	debugLog("Server count: " + strconv.Itoa(guildCount))
	s.UpdateStatus(0, "in " + strconv.Itoa(guildCount) + " servers!")

	guilds = make(map[string] string)

	for _, guildRow := range s.State.Guilds {
		guilds[guildRow.ID] = guildRow.Name
		debugLog(guildRow.ID + ": " + guildRow.Name)
	}
}

func guildDelete(s *discordgo.Session, event *discordgo.GuildDelete) {
	guildCount = len(s.State.Guilds)
	debugLog("Server count: " + strconv.Itoa(guildCount))
	s.UpdateStatus(0, "in " + strconv.Itoa(guildCount) + " servers!")

	guilds = make(map[string] string)

	for _, guildRow := range s.State.Guilds {
		guilds[guildRow.ID] = guildRow.Name
		debugLog(guildRow.ID + ": " + guildRow.Name)
	}
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Content == "" {
		return //No need to continue if there's no message
	}

	if (m.Author.ID == s.State.User.ID || m.Author.ID == "" || m.Author.Username == "") {
		return //Don't want the bot to reply to itself or to thin air
	}
	
	if m.ChannelID == "" {
		return //Where did this message even come from!?
	}
	
	contentWithMentionsReplaced := m.ContentWithMentionsReplaced()

	doesMessageExist := false
	for _, v := range messages {
		for obj := range v {
			if (obj.ChannelID == m.ChannelID && obj.ID == m.ID) {
				doesMessageExist = true
				break
			}
		}
		if doesMessageExist {
			break
		} else {
			return
		}
	}

	go handleMessage(s, m.Content, contentWithMentionsReplaced, m.Author.ID, m.Author.Username, m.Author.Discriminator, m.ChannelID, m.ID, true)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "" {
		return //No need to continue if there's no message
	}

	if (m.Author.ID == s.State.User.ID || m.Author.ID == "" || m.Author.Username == "") {
		return //Don't want the bot to reply to itself or to thin air
	}
	
	if m.ChannelID == "" {
		return //Where did this message even come from!?
	}
	
	contentWithMentionsReplaced := m.ContentWithMentionsReplaced()
	
	if messages[m.ChannelID] == nil {
		messages[m.ChannelID] = make(chan message, 0)
	}
	go func() {
		messages[m.ChannelID] <- message{ID:m.ID, ChannelID:m.ChannelID}
	}()

	go handleMessage(s, m.Content, contentWithMentionsReplaced, m.Author.ID, m.Author.Username, m.Author.Discriminator, m.ChannelID, m.ID, false)
}

func handleMessage(session *discordgo.Session, content string, contentWithMentionsReplaced string, authorID string, authorUsername string, authorDiscriminator string, channelID string, messageID string, updateMessage bool) {
	guildDetails, _ := guildDetails(channelID, session)
	channelDetails, _ := channelDetails(channelID, session)
	
	if guildDetails == nil || channelDetails == nil {
		return
	}

	debugLog("[" + guildDetails.Name + " #" + channelDetails.Name + "] " + authorUsername + "#" + authorDiscriminator + ": " + contentWithMentionsReplaced)
	
	if strings.HasPrefix(content, botPrefix) {
		c, err := session.State.Channel(channelID)
		if err != nil {
			// Could not find channel.
			return
		}
		g, err := session.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}
		
		session.ChannelTyping(channelID) // Send a typing event
		
		cmdMsg := strings.Replace(content, botPrefix, "", 1)
		cmd := strings.Split(cmdMsg, " ")
		switch cmd[0] {
			case "help":
				helpEmbed := NewEmbed().
					SetTitle(botName + " - Help").
					SetDescription("A list of available commands for " + botName + ".").
					AddField(botPrefix + "help", "Displays this help message.").
					AddField(botPrefix + "about", "Displays information about " + botName + " and how to use it.").
					AddField(botPrefix + "roll", "Rolls a dice.").
					AddField(botPrefix + "doubleroll", "Rolls two die.").
					AddField(botPrefix + "coinflip", "Flips a coin.").
					AddField(botPrefix + "xkcd (comic number|random|latest)", "Displays an xkcd comic depending on the requested type or comic number.").
					AddField(botPrefix + "imgur (url)", "Displays info about the specified Imgur image, album, gallery image, or gallery album.").
					AddField(botPrefix + "play (url)", "Plays the specified YouTube or direct audio URL in the user's current voice channel.").
					AddField(botPrefix + "youtube help", "Lists available YouTube commands.").
					AddField(botPrefix + "stop", "Stops the currently playing audio.").
					AddField(botPrefix + "leave", "Leaves the current voice channel.").
					SetColor(0xfafafa).MessageEmbed
				session.ChannelMessageSendEmbed(channelID, helpEmbed)
			case "about":
				aboutEmbed := NewEmbed().
					SetTitle(botName + " - About").
					SetDescription(botName + " is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
					AddField("How can I use " + botName + " in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
					AddField("How can I help keep " + botName + " running?", "The best ways to help keep " + botName + " running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
					AddField("How can I use " + botName + "?", "There are many ways to make use of " + botName + ".\n1) Type ``cli$help`` and try using some of the available commands.\n2) Ask " + botName + " a question, ex: ``Clinet, what time is it?`` or ``Clinet, what is DiscordApp?``.").
					AddField("Invite Link", "https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot").
					AddField("Donation Link", "https://www.paypal.me/JoshuaDoes").
					AddField("Source Code Link", "https://github.com/JoshuaDoes/clinet-discord/").
					SetColor(0x1c1c1c).MessageEmbed
				session.ChannelMessageSendEmbed(channelID, aboutEmbed)
			case "imgur":
				if len(cmd) > 1 {
					iErr := queryImgur(session, channelID, guildDetails.ID, cmd[1])
					if iErr {
						session.ChannelMessageSend(channelID, "Error finding info about the specified URL on Imgur.")
					}
				} else {
					session.ChannelMessageSend(channelID, "You must specify an Imgur URL to query Imgur.")
				}
			case "xkcd":
				if len(cmd) > 1 {
					switch cmd[1] {
						case "random":
							client := xkcd.NewClient()
							comic, err := client.Random()
							if err != nil {
								session.ChannelMessageSend(channelID, "Error finding random xkcd comic.")
								return
							}
							xkcdRandomEmbed := NewEmbed().
								SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
								SetDescription(comic.Title).
								SetImage(comic.ImageURL).
								SetColor(0x96a8c8).MessageEmbed
							session.ChannelMessageSendEmbed(channelID, xkcdRandomEmbed)
						case "latest":
							client := xkcd.NewClient()
							comic, err := client.Latest()
							if err != nil {
								session.ChannelMessageSend(channelID, "Error finding latest xkcd comic.")
								return
							}
							xkcdLatestEmbed := NewEmbed().
								SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
								SetDescription(comic.Title).
								SetImage(comic.ImageURL).
								SetColor(0x96a8c8).MessageEmbed
							session.ChannelMessageSendEmbed(channelID, xkcdLatestEmbed)
						default:
							client := xkcd.NewClient()
							comicNumber, err := strconv.Atoi(cmd[1])
							if err != nil {
								session.ChannelMessageSend(channelID, "``" + cmd[1] + "`` is not a valid number.")
							} else {
								comic, err := client.Get(comicNumber)
								if err != nil {
									session.ChannelMessageSend(channelID, "Error finding xkcd comic #" + cmd[1] + ".")
									return
								}
								xkcdSpecifiedEmbed := NewEmbed().
									SetTitle("xkcd - #" + cmd[1]).
									SetDescription(comic.Title).
									SetImage(comic.ImageURL).
									SetColor(0x96a8c8).MessageEmbed
								session.ChannelMessageSendEmbed(channelID, xkcdSpecifiedEmbed)
							}
					}
				} else {
					client := xkcd.NewClient()
					comic, err := client.Random()
					if err != nil {
						session.ChannelMessageSend(channelID, "Error finding random xkcd comic.")
					}
					xkcdRandomEmbed := NewEmbed().
						SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
						SetDescription(comic.Title).
						SetImage(comic.ImageURL).
						SetColor(0x96a8c8).MessageEmbed
					session.ChannelMessageSendEmbed(channelID, xkcdRandomEmbed)
				}
			case "play":
				go func() {
					for _, vs := range g.VoiceStates {
						if vs.UserID == authorID {
							if len(cmd) < 2 {
								err := playSound(session, g.ID, vs.ChannelID, channelID, "")
								if err != nil {
									debugLog("Error playing sound: " + fmt.Sprintf("%v", err))
									session.ChannelMessageSend(channelID, "Error playing sound.")
									return
								}
							} else {
								url := cmd[1]
								if url == "" {
									session.ChannelMessageSend(channelID, "You must specify a URL.")
									return
								}
								err := playSound(session, g.ID, vs.ChannelID, channelID, url)
								if err != nil {
									debugLog("Error playing sound: " + fmt.Sprintf("%v", err))
									session.ChannelMessageSend(channelID, "Error playing sound.")
									return
								}
							}
						}
					}
				}()
			case "stop":
				go func() {
					for _, vs := range g.VoiceStates {
						if vs.UserID == authorID {
							stopSound(g.ID, vs.ChannelID)
							session.ChannelMessageSend(channelID, "Stopped playing sound.")
							return
						}
					}
					session.ChannelMessageSend(channelID, "Error finding voice channel to stop audio playback in.")
				}()
			case "leave":
				go func() {
					for _, vs := range g.VoiceStates {
						if vs.UserID == authorID {
							voiceLeave(session, g.ID, vs.ChannelID)
							session.ChannelMessageSend(channelID, "Left voice channel.")
							return
						}
					}
					session.ChannelMessageSend(channelID, "Error finding voice channel to leave.")
				}()
			case "queue":
				viewQueue(session, g.ID, channelID)
			case "youtube":
				if len(cmd) < 2 {
					session.ChannelMessageSend(channelID, "You must specify a YouTube command.")
					return
				}
				switch cmd[1] {
					case "help":
						helpYouTubeEmbed := NewEmbed().
							SetTitle(botName + " - YouTube Help").
							SetDescription("A list of available YouTube commands for " + botName + ".").
							AddField(botPrefix + "youtube help", "Displays this YouTube help message.").
							AddField(botPrefix + "youtube search (query)", "Searches for the queried video and plays it in the user's current voice channel.").
							SetColor(0xff0000).MessageEmbed
						session.ChannelMessageSendEmbed(channelID, helpYouTubeEmbed)
					case "search":
						query := strings.Replace(content, botPrefix + "youtube search", "", -1)
						for {
							if strings.HasPrefix(query, " ") {
								query = strings.Replace(query, " ", "", 1)
							} else {
								break
							}
						}
						if query == "" {
							session.ChannelMessageSend(channelID, "You must specify a valid search query.")
							return
						}
						go func() {
							client := &http.Client{
								Transport: &transport.APIKey{Key: youtubeAPIKey},
							}
							service, err := youtube.New(client)
							if err != nil {
								session.ChannelMessageSend(channelID, "There was an error creating a new YouTube client.")
								return
							}
							call := service.Search.List("id,snippet").
									Q(query).
									MaxResults(50)
							response, err := call.Do()
							if err != nil {
								session.ChannelMessageSend(channelID, "There was an error searching YouTube for the specified query.")
								return
							}
							for _, item := range response.Items {
								switch item.Id.Kind {
									case "youtube#video":
										url := "https://youtube.com/watch?v=" + item.Id.VideoId
										for _, vs := range g.VoiceStates {
											if vs.UserID == authorID {
												err := playSound(session, g.ID, vs.ChannelID, channelID, url)
												if err != nil {
													debugLog("Error playing YouTube sound: " + fmt.Sprintf("%v", err))
													session.ChannelMessageSend(channelID, "There was an error playing the queried YouTube video.")
												}
											}
										}
										return
								}
							}
							session.ChannelMessageSend(channelID, "There was an error searching YouTube for the specified query.")
						}()
					default:
						session.ChannelMessageSend(channelID, "Unknown YouTube command. Type ``cli$youtube help`` for a list of YouTube commands.")
				}
			case "roll":
				random := rand.Intn(6) + 1
				session.ChannelMessageSend(channelID, "You rolled a " + strconv.Itoa(random) + "!")
			case "doubleroll":
				random1 := rand.Intn(6) + 1
				random2 := rand.Intn(6) + 1
				randomTotal := random1 + random2
				session.ChannelMessageSend(channelID, "You rolled a " + strconv.Itoa(random1) + " and a " + strconv.Itoa(random2) + ". The total is " + strconv.Itoa(randomTotal) + "!")
			case "coinflip":
				random := rand.Intn(3)
				switch random {
					case 0:
						session.ChannelMessageSend(channelID, "I flipped my coin... and it landed on heads!")
					case 1:
						session.ChannelMessageSend(channelID, "I flipped my coin... and it landed on tails!")
					default:
						session.ChannelMessageSend(channelID, "I flipped my coin... and it landed sideways!")
				}
			default:
				session.ChannelMessageSend(channelID, "Unknown command. Type ``cli$help`` for a list of commands.")
		}
	} else {
		go func() {
			regexpBotName, _ := regexp.MatchString("(.*?)" + botName + "(.*?)", content)
			if regexpBotName && strings.HasSuffix(content, "?") {
				session.ChannelTyping(channelID) // Send a typing event
				
				debugLog("### [START] Query")
				
				query := content
				debugLog("Original query: " + query)
				
				// Sanitize for Wolfram|Alpha
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
				debugLog("Sanitized query: " + query)
				
				iErr := queryDDG(channelID, session, query, messageID, guildDetails.ID)
				if iErr {
					iErr = queryWolfram(channelID, session, query, messageID, guildDetails.ID)
					if iErr {
						message, err := session.ChannelMessageSend(channelID, botName + " was unable to process your request.")
						if err == nil {
							responses[messageID] = message.ID
							session.MessageReactionAdd(channelID, messageID, "\u274C")
						} else {
							debugLog("Error sending message in [" + guildDetails.ID + ":" + channelID + "]")
						}
					}
				}
				
				debugLog("### [END]")
			}
		}()
	}
}

func queryWolfram(channelID string, session *discordgo.Session, query string, messageID string, guildID string) (bool) {
	queryResultObject, err := wolframClient.GetQueryResult(query, nil)
	if err != nil {
		debugLog(fmt.Sprintf("[Wolfram|Alpha] Error getting query result: %v", err))
		return true
	}
	
	queryResult := queryResultObject.QueryResult
	pods := queryResult.Pods
	
	if len(pods) < 1 {
		debugLog("[Wolfram|Alpha] Error getting pods from query")
		return true
	}

	fields := []*discordgo.MessageEmbedField{}
	embedImage := ""
	
	for podN, pod := range pods {
		podTitle := pod.Title
		switch podTitle {
			case "Locations":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
			case "Nearby locations":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
			case "Local map":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
			case "Inferred local map":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
			case "Inferred nearest city center":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
			case "IP address":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
			case "IP address registrant":
				debugLog("[Wolfram|Alpha] Denied pod: " + podTitle)
				continue
		}
		
		subPods := pod.SubPods
		if len(subPods) > 0 {
			debugLog("[Wolfram|Alpha] Pod #" + strconv.Itoa(podN + 1))
			for subPodN, subPod := range subPods {
				debugLog("[Wolfram|Alpha] Sub Pod #" + strconv.Itoa(subPodN + 1))

				plaintext := subPod.Plaintext
				imageSRC := subPod.Image.Src
				if plaintext != "" {
					// Make nicer for Discord
					plaintext = strings.Replace(plaintext, "Wolfram|Alpha", botName, -1)
					plaintext = strings.Replace(plaintext, "Wolfram Alpha", botName, -1)
					plaintext = strings.Replace(plaintext, "I was created by Stephen Wolfram and his team.", "I was created by JoshuaDoes.", -1)

					debugLog("Pod Title: " + podTitle)
					debugLog("Plaintext: " + plaintext)
					fields = append(fields, &discordgo.MessageEmbedField{Name:podTitle, Value:plaintext})
				}
				if imageSRC != "" && embedImage == "" && podTitle != "Input" && podTitle != "Input interpretation" {
					debugLog("Image SRC: " + imageSRC)
					embedImage = imageSRC
				}
			}
		}
	}
	
	resultEmbed := NewEmbed().
		SetColor(0xda0e1a).MessageEmbed
	
	if len(fields) == 0 {
		if embedImage != "" {
			resultEmbed.Image = &discordgo.MessageEmbedImage{URL:embedImage}
		} else {
			debugLog("[Wolfram|Alpha] Error getting legal data from available pods")
			return true
		}
	} else {
		resultEmbed.Fields = fields
		if embedImage != "" {
			resultEmbed.Image = &discordgo.MessageEmbedImage{URL:embedImage}
		}
	}
	message, err := session.ChannelMessageSendEmbed(channelID, resultEmbed)
	if err == nil {
		responses[messageID] = message.ID
		session.MessageReactionAdd(channelID, messageID, "\u2705")
	} else {
		debugLog("[Wolfram|Alpha] Error sending message in [" + guildID + ":" + channelID + "]")
		return true
	}
	return false
}

func queryDDG(channelID string, session *discordgo.Session, query string, messageID string, guildID string) (bool) {
	queryResult, err := ddgClient.GetQueryResult(query)
	if err != nil {
		debugLog(fmt.Sprintf("[DuckDuckGo] Error getting query result: %v", err))
		return true
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
		debugLog("[DuckDuckGo] Error getting query result from response")
		return true
	}

	resultEmbed := NewEmbed().
		SetTitle(queryResult.Heading).
		SetDescription(result).
		SetColor(0xdf5730).MessageEmbed
	if queryResult.Image != "" {
		resultEmbed.Image = &discordgo.MessageEmbedImage{URL:queryResult.Image}
	}
	message, err := session.ChannelMessageSendEmbed(channelID, resultEmbed)
	if err == nil {
		responses[messageID] = message.ID
		session.MessageReactionAdd(channelID, messageID, "\u2705")
	} else {
		debugLog("[DuckDuckGo] Error sending message in [" + guildID + ":" + channelID + "]")
		return true
	}
	return false
}

func queryImgur(session *discordgo.Session, channelID, guildID, url string) (bool) {
	imgurInfo, _, err := imgurClient.GetInfoFromURL(url)
	if err != nil {
		debugLog("[Imgur] Error getting info from URL [" + url + "]")
		return true
	}
	if imgurInfo.Image != nil {
		debugLog("[Imgur] Detected image from URL [" + url + "]")
		imgurImage := imgurInfo.Image
		imgurEmbed := NewEmbed().
			SetTitle(imgurImage.Title).
			SetDescription(imgurImage.Description).
			AddField("Views", strconv.Itoa(imgurImage.Views)).
			AddField("NSFW", strconv.FormatBool(imgurImage.Nsfw)).
			SetColor(0x89c623).MessageEmbed
		_, err := session.ChannelMessageSendEmbed(channelID, imgurEmbed)
		if err != nil {
			debugLog("[Imgur] Error sending message in [" + guildID + ":" + channelID + "]")
			return true
		}
	} else if imgurInfo.Album != nil {
		debugLog("[Imgur] Detected album from URL [" + url + "]")
		imgurAlbum := imgurInfo.Album
		imgurEmbed := NewEmbed().
			SetTitle(imgurAlbum.Title).
			SetDescription(imgurAlbum.Description).
			AddField("Uploader", imgurAlbum.AccountURL).
			AddField("Image Count", strconv.Itoa(imgurAlbum.ImagesCount)).
			AddField("Views", strconv.Itoa(imgurAlbum.Views)).
			AddField("NSFW", strconv.FormatBool(imgurAlbum.Nsfw)).
			SetColor(0x89c623).MessageEmbed
		_, err = session.ChannelMessageSendEmbed(channelID, imgurEmbed)
		if err != nil {
			debugLog("[Imgur] Error sending message in [" + guildID + ":" + channelID + "]")
			return true
		}
	} else if imgurInfo.GImage != nil {
		debugLog("[Imgur] Detected gallery image from URL [" + url + "]")
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
		_, err := session.ChannelMessageSendEmbed(channelID, imgurEmbed)
		if err != nil {
			debugLog("[Imgur] Error sending message in [" + guildID + ":" + channelID + "]")
			return true
		}
	} else if imgurInfo.GAlbum != nil {
		debugLog("[Imgur] Detected gallery album from URL [" + url + "]")
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
		_, err = session.ChannelMessageSendEmbed(channelID, imgurEmbed)
		if err != nil {
			debugLog("[Imgur] Error sending message in [" + guildID + ":" + channelID + "]")
			return true
		}
	} else {
		debugLog("[Imgur] Error detecting Imgur type from URL [" + url + "]")
		return true
	}
	return false
}

func guildDetails(channelID string, s *discordgo.Session) (*discordgo.Guild, error) {
	channelInGuild, err := s.State.Channel(channelID)
	if err != nil {
		return nil, err
	}
	guildDetails, err := s.State.Guild(channelInGuild.GuildID)
	if err != nil {
		return nil, err
	}
	return guildDetails, nil
}

func channelDetails(channelID string, s *discordgo.Session) (*discordgo.Channel, error) {
	channelInGuild, err := s.State.Channel(channelID)
	if err != nil {
		return nil, err
	}
	return channelInGuild, nil
}

func clearVoiceSession(guildID string) {
	queue[guildID] = &GuildQueue{}
	voiceData[guildID] = &VoiceData{}
}

func voiceLeave(s *discordgo.Session, guildID, channelID string) {
	for _, voiceDataRow := range voiceData {
		if voiceDataRow.VoiceConnection.ChannelID == channelID {
			debugLog("A> Leaving voice channel [" + guildID + ":" + channelID + "]...")
			voiceDataRow.IsPlaybackRunning = false
			voiceDataRow.WasPlaybackStoppedManually = false
			voiceDataRow.VoiceConnection.Disconnect()
			
			clearVoiceSession(guildID)
			
			return
		}
	}
}

func stopSound(guildID, channelID string) {
	for _, voiceDataRow := range voiceData {
		if voiceDataRow.VoiceConnection.ChannelID == channelID {
			debugLog("A> Stopping sound on voice channel [" + guildID + ":" + channelID + "]...")
			voiceDataRow.IsPlaybackRunning = false
			voiceDataRow.WasPlaybackStoppedManually = true
			
			return
		}
	}
}

func playSound(s *discordgo.Session, guildID, channelID string, callerChannelID string, url string) (err error) {
	var voiceConnection *discordgo.VoiceConnection = nil
	var encodingSession *dca.EncodeSession = nil
	var stream *dca.StreamingSession = nil
	var isPlaybackRunning bool = false
	for _, voiceDataRow := range voiceData {
		if voiceDataRow.VoiceConnection != nil {
			if voiceDataRow.VoiceConnection.ChannelID == channelID {
				debugLog("A> Found previous connection to voice channel [" + guildID + ":" + channelID + "]")
				voiceConnection = voiceDataRow.VoiceConnection
				encodingSession = voiceDataRow.EncodingSession
				stream = voiceDataRow.Stream
				isPlaybackRunning = voiceDataRow.IsPlaybackRunning
				break
			}
		}
	}

	if voiceConnection == nil {
		debugLog("1B> Connecting to voice channel [" + guildID + ":" + channelID + "]...")
		voiceConnection, err = s.ChannelVoiceJoin(guildID, channelID, false, false)
		if err != nil {
			debugLog("1C> Error connecting to voice channel [" + guildID + ":" + channelID + "]")
			voiceConnection.Disconnect()
			return err
		}
	}
	
	_, ok := queue[guildID]
	if ok {
		debugLog("Guild queue previously initialized")
		if isPlaybackRunning {
			debugLog("Playback in progress, appending to guild queue...")
			
			title := ""
			author := ""
			imageURL := ""
			thumbnailURL := ""
			regexpHasYouTube, _ := regexp.MatchString("(?:https?:\\/\\/)?(?:www\\.)?youtu\\.?be(?:\\.com)?\\/?.*(?:watch|embed)?(?:.*v=|v\\/|\\/)(?:[\\w-_]+)", url)
			if regexpHasYouTube {
				videoInfo, err := ytdl.GetVideoInfo(url)
				if err != nil {
					return err
				}
				title = videoInfo.Title
				author = videoInfo.Author
				imageURL = videoInfo.GetThumbnailURL("maxresdefault").String()
				thumbnailURL = videoInfo.GetThumbnailURL("default").String()
				
				format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
				_, err = videoInfo.GetDownloadURL(format)
				if err != nil {
					return err
				}
			}
			
			newEntry := &Queue{Name:title, Author:author, ImageURL:imageURL, ThumbnailURL:thumbnailURL, URL:url}
			queue[guildID].Queue = append(queue[guildID].Queue, *newEntry)
			debugLog(fmt.Sprintf("%v", queue))
			if regexpHasYouTube {
				embed := NewEmbed().
					SetTitle("Added to Queue").
					AddField(title, author).
					//SetImage(imageURL).
					SetThumbnail(thumbnailURL).
					SetColor(0xff0000).MessageEmbed
				s.ChannelMessageSendEmbed(callerChannelID, embed)
			} else {
				s.ChannelMessageSend(callerChannelID, "Added ``" + title + "`` to the queue.")
			}
			return
		} else {
			debugLog("Continuing with playback")
			if url == "" {
				if len(queue[guildID].Queue) > 0 {
					debugLog("Queued URL found in guild queue, fetching URL...")
					url = queue[guildID].Queue[0].URL
					debugLog("Removing queued URL from guild queue...")
					queue[guildID].Queue[len(queue[guildID].Queue) - 1], queue[guildID].Queue[0] = queue[guildID].Queue[0], queue[guildID].Queue[len(queue[guildID].Queue) - 1]
					queue[guildID].Queue = queue[guildID].Queue[:len(queue[guildID].Queue) - 1]
					debugLog("Current guild queue: " + fmt.Sprintf("%v", queue))
					debugLog("Playing URL [" + url + "] from guild queue...")
					playSound(s, guildID, channelID, callerChannelID, url)
					return
				} else {
					debugLog("No entries left in the guild queue")
					s.ChannelMessageSend(callerChannelID, "No entries were found in the guild queue.")
					return errors.New("Unable to find an entry in the guild queue")
				}
			}
		}
	} else {
		debugLog("Initializing guild queue...")
		queue[guildID] = &GuildQueue{}
		debugLog(fmt.Sprintf("%v", queue))
		debugLog("Continuing with playback")
	}

	debugLog("1D> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
	voiceConnection.Speaking(false)
	
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"
	
	mediaURL := url
	var embedMessage *discordgo.Message
	embedMessageID := ""
	title := ""
	author := ""
	//imageURL := ""
	thumbnailURL := ""
	
	regexpHasYouTube, _ := regexp.MatchString("(?:https?:\\/\\/)?(?:www\\.)?youtu\\.?be(?:\\.com)?\\/?.*(?:watch|embed)?(?:.*v=|v\\/|\\/)(?:[\\w-_]+)", url)
	if regexpHasYouTube {
		videoInfo, err := ytdl.GetVideoInfo(url)
		if err != nil {
			debugLog("1E> Error getting video info from [" + url + "]")
			return err
		}
		
		debugLog("1F> Storing video metadata...")
		title = videoInfo.Title
		author = videoInfo.Author
		//imageURL = videoInfo.GetThumbnailURL("maxresdefault").String()
		thumbnailURL = videoInfo.GetThumbnailURL("default").String()
		
		format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
		downloadURL, err := videoInfo.GetDownloadURL(format)
		if err != nil {
			debugLog("1G> Error getting download URL from [" + url + "]")
			return err
		}
		mediaURL = downloadURL.String()
		
		embed := NewEmbed().
			SetTitle(title).
			SetDescription(author).
			AddField("Duration", "0s").
			//SetImage(imageURL).
			SetThumbnail(thumbnailURL).
			SetColor(0xff0000).MessageEmbed
		embedMessage, _ = s.ChannelMessageSendEmbed(callerChannelID, embed)
		embedMessageID = embedMessage.ID
	} else {
		embed := NewEmbed().
			AddField("URL", mediaURL).
			AddField("Duration", "0s").
			SetColor(0xffffff).MessageEmbed
		embedMessage, _ = s.ChannelMessageSendEmbed(callerChannelID, embed)
		embedMessageID = embedMessage.ID
	}
	
	encodingSession, err = dca.EncodeFile(mediaURL, options)
	if err != nil {
		debugLog("1I> Error encoding file [" + mediaURL + "]")
		return err
	}
    
	debugLog("1K> Setting speaking to true in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
	voiceConnection.Speaking(true)
    
	done := make(chan error)
	stream = dca.NewStream(encodingSession, voiceConnection, done)

	debugLog("1L> Storing voiceConnection, encodingSession, stream, and playbackRunning handles/states in memory...")
	voiceData[guildID] = &VoiceData{VoiceConnection:voiceConnection, EncodingSession:encodingSession, Stream:stream, IsPlaybackRunning:true, WasPlaybackStoppedManually:false}
	isPlaybackRunning = true

	ticker := time.NewTicker(time.Second)
	
	for voiceData[guildID].IsPlaybackRunning {
		select {
			case err := <- done:
				if err != nil {
					fmt.Println("Playback finished")
					voiceData[guildID].IsPlaybackRunning = false
					break
				} else {
					fmt.Println("Playback not finished")
				}
			case <- ticker.C:
				duration := Round(stream.PlaybackPosition(), time.Second)
				if regexpHasYouTube {
					embed := NewEmbed().
						SetTitle(title).
						SetDescription(author).
						AddField("Duration", duration.String()).
						//SetImage(imageURL).
						SetThumbnail(thumbnailURL).
						SetColor(0xff0000).MessageEmbed
					s.ChannelMessageEditEmbed(callerChannelID, embedMessageID, embed)
				} else {
					embed := NewEmbed().
						AddField("URL", mediaURL).
						AddField("Duration", duration.String()).
						SetColor(0xffffff).MessageEmbed
					s.ChannelMessageEditEmbed(callerChannelID, embedMessageID, embed)
				}
		}
	}
	
	debugLog("1T> Cleaning up encoding session...")
	encodingSession.Stop()
	encodingSession.Cleanup()
	encodingSession.Truncate()
    
	debugLog("1U> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
	voiceConnection.Speaking(false)
	
	isPlaybackRunning = false
	ticker.Stop()
	
	if len(queue[guildID].Queue) == 0 {
		debugLog("Guild queue empty, leaving voice channel...")
		voiceLeave(s, guildID, channelID)
	} else {
		if voiceData[guildID].WasPlaybackStoppedManually == false {
			debugLog("Queued URL found in guild queue, fetching URL...")
			url = queue[guildID].Queue[0].URL
			debugLog("Removing queued URL from guild queue...")
			queue[guildID].Queue[len(queue[guildID].Queue) - 1], queue[guildID].Queue[0] = queue[guildID].Queue[0], queue[guildID].Queue[len(queue[guildID].Queue) - 1]
			queue[guildID].Queue = queue[guildID].Queue[:len(queue[guildID].Queue) - 1]
			debugLog("Current guild queue: " + fmt.Sprintf("%v", queue))
			debugLog("Playing URL [" + url + "] from guild queue...")
			playSound(s, guildID, channelID, callerChannelID, url)
		} else {
			debugLog("Was told to stop, staying...")
		}
	}
	
	return nil
}

func viewQueue(session *discordgo.Session, guildID, channelID string) {
	guildQueue := queue[guildID].Queue
	if len(guildQueue) > 0 {
		queueList := ""
		count := 0
		for _, queueRow := range guildQueue {
			count += 1
			if queueList == "" {
				queueList = "``" + strconv.Itoa(count) + ".`` " + queueRow.URL
			} else {
				queueList = queueList + "\n``" + strconv.Itoa(count) + ".`` " + queueRow.URL
			}
		}
		guildDetails, _ := guildDetails(channelID, session)
		queueEmbed := NewEmbed().
			SetTitle("Queue for ``" + guildDetails.Name + "``").
			SetDescription(queueList).
			SetColor(0xfafafa).MessageEmbed
		_, err := session.ChannelMessageSendEmbed(channelID, queueEmbed)
		if err != nil {
			debugLog("Error sending message in [" + guildID + ":" + channelID + "]")
		}
	} else {
		session.ChannelMessageSend(channelID, "There are no entries in the queue.")
	}
}

func Round(d, r time.Duration) time.Duration {
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