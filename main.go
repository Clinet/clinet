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

	"github.com/paked/configure" // Allows configuration of the program via external sources
	"github.com/bwmarrin/discordgo" // Allows usage of the Discord API
	"github.com/JoshuaDoes/go-wolfram" // Allows usage of the Wolfram|Alpha API
	"github.com/jonas747/dca" // Allows the encoding/decoding of the Discord Audio format
	"github.com/rylio/ytdl" // Allows the fetching of YouTube video metadata and download URLs
	"google.golang.org/api/googleapi/transport" // Allows the making of authenticated API requests to Google
	"google.golang.org/api/youtube/v3" // Allows usage of the YouTube API
	"github.com/nishanths/go-xkcd" // Allows the fetching of XKCD comics
)

type message struct {
    ID  string
    ChannelID string   
}

var (
	conf = configure.New()
	confBotToken = conf.String("botToken", "", "Bot Token")
	confBotName = conf.String("botName", "", "Bot Name")
	confBotPrefix = conf.String("botPrefix", "", "Bot Prefix")
	confWolframAppID = conf.String("wolframAppID", "", "Wolfram AppID")
	confYouTubeAPIKey = conf.String("youtubeAPIKey", "", "YouTube API Key")
	confDebugMode = conf.Bool("debugMode", false, "Debug Mode")
	botToken string = ""
	botName string = ""
	botPrefix string = ""
	wolframAppID string = ""
	youtubeAPIKey string = ""
	debugMode bool = false
	
	wolframClient *wolfram.Client
	
	guildCount int
	guilds = make(map[string] string)
	
	voiceConnections []*discordgo.VoiceConnection
	encodingSessions []*dca.EncodeSession
	streams []*dca.StreamingSession
	playbackRunning []bool
	queue = make(map[string] []string)
	
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
	youtubeAPIKey = *confYouTubeAPIKey
	debugMode = *confDebugMode
	if (botToken == "" || botName == "" || botPrefix == "" || wolframAppID == "" || youtubeAPIKey == "") {
		fmt.Println("> Configuration not properly setup, exiting...")
		return
	} else {
		fmt.Println("> Successfully loaded configuration.")
		debugLog("botToken: " + botToken)
		debugLog("botName: " + botName)
		debugLog("botPrefix: " + botPrefix)
		debugLog("wolframAppID: " + wolframAppID)
		debugLog("youtubeAPIKey: " + youtubeAPIKey)
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

	fmt.Println("> " + botName + " has started successfully.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if len(voiceConnections) > 0 {
		debugLog("> Closing any active voice connections...")
		for _, voiceConnectionRow := range voiceConnections {
			voiceConnectionRow.Close()
		}
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

	handleMessage(s, m.Content, contentWithMentionsReplaced, m.Author.ID, m.Author.Username, m.Author.Discriminator, m.ChannelID, m.ID, true)
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

	handleMessage(s, m.Content, contentWithMentionsReplaced, m.Author.ID, m.Author.Username, m.Author.Discriminator, m.ChannelID, m.ID, false)
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
					AddField(botPrefix + "xkcd (comic number|random|latest)", "Displays an xkcd comic depending on the requested type or comic number.").
					AddField(botPrefix + "play (url)", "Plays the specified YouTube or direct audio URL in the user's current voice channel.").
					AddField(botPrefix + "youtube help", "Lists available YouTube commands.").
					AddField(botPrefix + "stop", "Stops the currently playing audio.").
					AddField(botPrefix + "leave", "Leaves the current voice channel.").
					SetColor(0xfafafa).MessageEmbed
				session.ChannelMessageSendEmbed(channelID, helpEmbed)
			case "xkcd":
				if len(cmd) > 1 {
					switch cmd[1] {
						case "random":
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
						case "latest":
							client := xkcd.NewClient()
							comic, err := client.Latest()
							if err != nil {
								session.ChannelMessageSend(channelID, "Error finding latest xkcd comic.")
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
				if len(cmd) < 2 {
					session.ChannelMessageSend(channelID, "You must specify a valid URL.")
					return
				}
				url := cmd[1]
				if url == "" {
					session.ChannelMessageSend(channelID, "You must specify a valid URL.")
					return
				}
				for _, vs := range g.VoiceStates {
					if vs.UserID == authorID {
						err := playSound(session, g.ID, vs.ChannelID, channelID, url)
						if err != nil {
							debugLog("Error playing sound:" + fmt.Sprintf("%v", err))
							session.ChannelMessageSend(channelID, "Error playing sound.")
							return
						}
					}
				}
			case "stop":
				for _, vs := range g.VoiceStates {
					if vs.UserID == authorID {
						stopSound(g.ID, vs.ChannelID)
						session.ChannelMessageSend(channelID, "Stopped playing sound.")
						return
					}
				}
				session.ChannelMessageSend(channelID, "Error finding voice channel to stop audio playback in.")
			case "leave":
				for _, vs := range g.VoiceStates {
					if vs.UserID == authorID {
						voiceLeave(session, g.ID, vs.ChannelID)
						session.ChannelMessageSend(channelID, "Left voice channel.")
						return
					}
				}
				session.ChannelMessageSend(channelID, "Error finding voice channel to leave.")
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
					default:
						session.ChannelMessageSend(channelID, "Unknown YouTube command. Type ``cli$youtube help`` for a list of YouTube commands.")
				}
			default:
				session.ChannelMessageSend(channelID, "Unknown command. Type ``cli$help`` for a list of commands.")
		}
	} else {
		regexpBotName, _ := regexp.MatchString("(.*?)" + botName + "(.*?)", content)
		if regexpBotName && strings.HasSuffix(content, "?") {
			session.ChannelTyping(channelID) // Send a typing event
			
			debugLog("### [START] Wolfram")
			
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
			
			queryResultObject, err := wolframClient.GetQueryResult(query, nil)
			if err != nil {
				message, err := session.ChannelMessageSend(channelID, botName + " was unable to process your request.")
				if err == nil {
					responses[messageID] = message.ID
					session.MessageReactionAdd(channelID, messageID, "\u274C")
				} else {
					debugLog("Error sending message in [" + guildDetails.ID + ":" + channelID + "]")
				}
				debugLog(fmt.Sprintf("Error getting query result: %v", err))
				return
			}
			
			queryResult := queryResultObject.QueryResult
			pods := queryResult.Pods
			
			if len(pods) < 1 {
				message, err := session.ChannelMessageSend(channelID, botName + " was unable to process your request.")
				if err == nil {
					responses[messageID] = message.ID
					session.MessageReactionAdd(channelID, messageID, "\u274C")
				} else {
					debugLog("Error sending message in [" + guildDetails.ID + ":" + channelID + "]")
				}
				debugLog("Error getting pods from query")
				return
			}
			
			result := ""
			
			for _, pod := range pods {
				podTitle := pod.Title
				switch podTitle {
					case "Locations":
						debugLog("Denied pod: " + podTitle)
						continue
					case "Nearby locations":
						debugLog("Denied pod: " + podTitle)
						continue
					case "Local map":
						debugLog("Denied pod: " + podTitle)
						continue
					case "Inferred local map":
						debugLog("Denied pod: " + podTitle)
						continue
					case "Inferred nearest city center":
						debugLog("Denied pod: " + podTitle)
						continue
					case "IP address":
						debugLog("Denied pod: " + podTitle)
						continue
					case "IP address registrant":
						debugLog("Denied pod: " + podTitle)
						continue
				}
				
				subPods := pod.SubPods
				if len(subPods) > 0 {
					for _, subPod := range subPods {
						plaintext := subPod.Plaintext
						if plaintext != "" {
							debugLog("Found result from pod [" + podTitle + "]: " + plaintext)
							if result != "" {
								result = result + "\n\n[" + podTitle + "]\n" + plaintext
							} else {
								result = "[" + podTitle + "]\n" + plaintext
							}
						}
					}
				}
			}
			
			if result == "" {
				message, err := session.ChannelMessageSend(channelID, botName + " was either unable to process your request or was denied permission from doing so.")
				if err == nil {
					responses[messageID] = message.ID
					session.MessageReactionAdd(channelID, messageID, "\u274C")
				} else {
					debugLog("Error sending message in [" + guildDetails.ID + ":" + channelID + "]")
				}
				debugLog("Error getting legal data from available pods")
				return
			}
			
			// Make nicer for Discord
			result = strings.Replace(result, "Wolfram|Alpha", botName, -1)
			result = strings.Replace(result, "Wolfram Alpha", botName, -1)
			result = strings.Replace(result, "I was created by Stephen Wolfram and his team.", "I was created by JoshuaDoes.", -1)
			
			if updateMessage {
				message, _ := session.ChannelMessageEdit(channelID, responses[messageID], result)
				if err == nil {
					responses[messageID] = message.ID
					session.MessageReactionAdd(channelID, messageID, "\u2705")
				} else {
					debugLog("Error updating message [" + messageID + "] in [" + guildDetails.ID + ":" + channelID + "]")
				}
			} else {
				message, err := session.ChannelMessageSend(channelID, result)
				if err == nil {
					responses[messageID] = message.ID
					session.MessageReactionAdd(channelID, messageID, "\u2705")
				} else {
					debugLog("Error sending message in [" + guildDetails.ID + ":" + channelID + "]")
				}
			}
			
			debugLog("### [END]")
		}
	}
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

func clearVoiceSession(i int) {
	voiceConnections[len(voiceConnections) - 1], voiceConnections[i] = voiceConnections[i], voiceConnections[len(voiceConnections) - 1]
	voiceConnections = voiceConnections[:len(voiceConnections) - 1]			
	encodingSessions[len(encodingSessions) - 1], encodingSessions[i] = encodingSessions[i], encodingSessions[len(encodingSessions) - 1]
	encodingSessions = encodingSessions[:len(encodingSessions) - 1]			
	streams[len(streams) - 1], streams[i] = streams[i], streams[len(streams) - 1]
	streams = streams[:len(streams) - 1]
}

func voiceLeave(s *discordgo.Session, guildID, channelID string) {
	for i, voiceConnectionRow := range voiceConnections {
		if voiceConnectionRow.ChannelID == channelID {
			debugLog("A> Leaving voice channel [" + guildID + ":" + channelID + "]...")
			playbackRunning[i] = false
			voiceConnectionRow.Disconnect()
			
			clearVoiceSession(i)
			
			return
		}
	}
}

func stopSound(guildID, channelID string) {
	for i, voiceConnectionRow := range voiceConnections {
		if voiceConnectionRow.ChannelID == channelID {
			debugLog("A> Stopping sound on voice channel [" + guildID + ":" + channelID + "]...")
			playbackRunning[i] = false
			return
		}
	}
}

func playSound(s *discordgo.Session, guildID, channelID string, callerChannelID string, url string) (err error) {
	var voiceConnection *discordgo.VoiceConnection = nil
	var encodingSession *dca.EncodeSession = nil
	var stream *dca.StreamingSession = nil
	var index int = -1
	var newRows bool = true
	for i, voiceConnectionRow := range voiceConnections {
		if voiceConnectionRow.ChannelID == channelID {
			debugLog("A> Found previous connection to voice channel [" + guildID + ":" + channelID + "]")
			voiceConnection = voiceConnections[i]
			encodingSession = encodingSessions[i]
			stream = streams[i]
			playbackRunning[i] = true
			index = i
			newRows = false
			break
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
	
	guildQueue, ok := queue[guildID]
	if ok {
		fmt.Println("Guild queue previously initialized")
		if stream != nil {
			fmt.Println("Playback in progress, appending to guild queue...")
			guildQueue = append(guildQueue, url)
			queue[guildID] = guildQueue
			fmt.Println(fmt.Sprintf("%v", queue))
			return
		}
	} else {
		fmt.Println("Initializing guild queue...")
		guildQueue = []string{}
		queue[guildID] = guildQueue
		fmt.Println(fmt.Sprintf("%v", queue))
		if stream != nil {
			fmt.Println("Playback in progress, appending to guild queue...")
			guildQueue = append(guildQueue, url)
			queue[guildID] = guildQueue
			fmt.Println(fmt.Sprintf("%v", queue))
			return
		}
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
	
	if newRows {
		debugLog("1L> Storing voiceConnection, encodingSession, stream, and playbackRunning handles/states in memory...")
		voiceConnections = append(voiceConnections, voiceConnection)
		encodingSessions = append(encodingSessions, encodingSession)
		streams = append(streams, stream)
		playbackRunning = append(playbackRunning, true)
		index = len(playbackRunning) - 1
	}
	
	ticker := time.NewTicker(time.Second)
	playbackFinished := false
	
	for playbackRunning[index] {
		select {
			case err := <- done:
				if err != nil {
					fmt.Println("Playback finished")
					playbackRunning[index] = false
					playbackFinished = true
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
	
	debugLog("1U> Writing nil to encoding session and stream...")
	encodingSession = nil
	encodingSessions[index] = nil
	stream = nil
	streams[index] = nil
    
	debugLog("1V> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
	voiceConnection.Speaking(false)
	
	ticker.Stop()
	
	if len(queue[guildID]) == 0 {
		fmt.Println("Guild queue empty, leaving voice channel...")
		voiceLeave(s, guildID, channelID)
	} else {
		if playbackFinished {
			fmt.Println("Queued URL found in guild queue, fetching URL...")
			url = queue[guildID][0]
			fmt.Println("Removing queued URL from guild queue...")
			queue[guildID][len(queue[guildID]) - 1], queue[guildID][0] = queue[guildID][0], queue[guildID][len(queue[guildID]) - 1]
			queue[guildID] = queue[guildID][:len(queue[guildID]) - 1]
			fmt.Println("Current guild queue: " + fmt.Sprintf("%v", queue))
			fmt.Println("Playing URL [" + url + "] from guild queue...")
			playSound(s, guildID, channelID, callerChannelID, url)
		} else {
			fmt.Println("Playback not marked as finished, doing nothing...")
		}
	}
	
	return nil
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