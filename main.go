package main

import (
	"fmt"
	"io"
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
)

var (
	conf = configure.New()
	confBotToken = conf.String("botToken", "", "Bot Token")
	confBotName = conf.String("botName", "", "Bot Name")
	confBotPrefix = conf.String("botPrefix", "", "Bot Prefix")
	confWolframAppID = conf.String("wolframAppID", "", "Wolfram AppID")
	confYouTubeAPIKey = conf.String("youtubeAPIKey", "", "YouTube API Key")
	botToken string = ""
	botName string = ""
	botPrefix string = ""
	wolframAppID string = ""
	youtubeAPIKey string = ""

	Token string
	
	wolframClient *wolfram.Client
	
	guildCount int
	
	voiceConnections []*discordgo.VoiceConnection
	encodingSessions []*dca.EncodeSession
	streams []*dca.StreamingSession
	playbackStopped []bool
)

func init() {
	conf.Use(configure.NewFlag())
	conf.Use(configure.NewJSONFromFile("config.json"))
}

func main() {
	fmt.Println("> Loading configuration...")
	conf.Parse()
	botToken = *confBotToken
	botName = *confBotName
	botPrefix = *confBotPrefix
	wolframAppID = *confWolframAppID
	youtubeAPIKey = *confYouTubeAPIKey
	if (botToken == "" || botName == "" || botPrefix == "" || wolframAppID == "" || youtubeAPIKey == "") {
		fmt.Println("> Configuration not properly setup, exiting...")
		return
	} else {
		fmt.Println("> Successfully loaded configuration.")
		fmt.Println("botToken: " + botToken)
		fmt.Println("botName: " + botName)
		fmt.Println("botPrefix: " + botPrefix)
		fmt.Println("wolframAppID: " + wolframAppID)
		fmt.Println("youtubeAPIKey: " + youtubeAPIKey)
	}
	
	fmt.Println("> Creating a new Discord session...")
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	
	fmt.Println("> Registering Ready callback handler...")
	dg.AddHandler(ready)

	fmt.Println("> Registering MessageCreate callback handler...")
	dg.AddHandler(messageCreate)
	
	fmt.Println("> Registering GuildJoin callback handler...")
	dg.AddHandler(guildCreate)

	fmt.Println("> Establishing a websocket connection to Discord...")
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}
	
	fmt.Println("> Initializing Wolfram...")
	wolframClient = &wolfram.Client{AppID:wolframAppID}

	fmt.Println("> " + botName + " has started successfully.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	for _, voiceConnectionRow := range voiceConnections {
		voiceConnectionRow.Close()
	}
	
	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	guildCount = len(s.State.Guilds)
	s.UpdateStatus(0, "in " + strconv.Itoa(guildCount) + " servers!")
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	guildCount = len(s.State.Guilds)
	s.UpdateStatus(0, "in " + strconv.Itoa(guildCount) + " servers!")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "" {
		return //No need to continue if there's no message
	}

	if m.Author.ID == s.State.User.ID {
		return //Don't want the bot to reply to itself
	}

	guildDetails, _ := guildDetails(m.ChannelID, s)
	channelDetails, _ := channelDetails(m.ChannelID, s)

	fmt.Println("[" + guildDetails.Name + " #" + channelDetails.Name + "] " + m.Author.Username + "#" + m.Author.Discriminator + ": " + m.ContentWithMentionsReplaced())
	
	if strings.HasPrefix(m.Content, botPrefix) {
		s.ChannelTyping(m.ChannelID) // Send a typing event

		if strings.HasPrefix(m.Content, botPrefix + "play ") {
			url := strings.Replace(m.Content, botPrefix + "play ", "", -1)
			if url == "" {
				s.ChannelMessageSend(m.ChannelID, "You must specify a valid URL.")
				return
			}
			c, err := s.State.Channel(m.ChannelID)
			if err != nil {
				// Could not find channel.
				return
			}
			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				// Could not find guild.
				return
			}
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Author.ID {
					err := playSound(s, g.ID, vs.ChannelID, m.ChannelID, url)
					if err != nil {
						fmt.Println("Error playing sound:", err)
						s.ChannelMessageSend(m.ChannelID, "Error playing sound.")
						return
					}
				}
			}
		} else if strings.HasPrefix(m.Content, botPrefix + "stop") {
			c, err := s.State.Channel(m.ChannelID)
			if err != nil {
				// Could not find channel.
				return
			}
			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				// Could not find guild.
				return
			}
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Author.ID {
					stopSound(g.ID, vs.ChannelID)
					s.ChannelMessageSend(m.ChannelID, "Stopped playing sound.")
				}
			}
		} else if strings.HasPrefix(m.Content, botPrefix + "leave") {
			c, err := s.State.Channel(m.ChannelID)
			if err != nil {
				// Could not find channel.
				return
			}
			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				// Could not find guild.
				return
			}
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Author.ID {
					voiceLeave(s, g.ID, vs.ChannelID)
					s.ChannelMessageSend(m.ChannelID, "Left voice channel.")
				}
			}
		} else if strings.HasPrefix(m.Content, botPrefix + "youtube search ") {
			c, err := s.State.Channel(m.ChannelID)
			if err != nil {
				// Could not find channel.
				return
			}
			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				// Could not find guild.
				return
			}
			query := strings.Replace(m.Content, botPrefix + "youtube search ", "", -1)
			if query == "" {
				s.ChannelMessageSend(m.ChannelID, "You must specify a valid search query.")
				return
			}
			client := &http.Client{
				Transport: &transport.APIKey{Key: youtubeAPIKey},
			}
			service, err := youtube.New(client)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "There was an error creating a new YouTube client.")
				return
			}
			call := service.Search.List("id,snippet").
					Q(query).
					MaxResults(50)
			response, err := call.Do()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "There was an error searching YouTube for the specified query.")
				return
			}
			for _, item := range response.Items {
				switch item.Id.Kind {
					case "youtube#video":
						url := "https://youtube.com/watch?v=" + item.Id.VideoId
						for _, vs := range g.VoiceStates {
							if vs.UserID == m.Author.ID {
								err := playSound(s, g.ID, vs.ChannelID, m.ChannelID, url)
								if err != nil {
									fmt.Println("Error playing YouTube sound:", err)
									s.ChannelMessageSend(m.ChannelID, "There was an error playing the queried YouTube video.")
								}
							}
						}
						return
				}
			}
			s.ChannelMessageSend(m.ChannelID, "There was an error searching YouTube for the specified query.")
		} else {
			s.ChannelMessageSend(m.ChannelID, "Unknown command.")
		}
	} else {
		regexpBotName, _ := regexp.MatchString("(.*?)" + botName + "(.*?)", m.Content)
		if regexpBotName && strings.HasSuffix(m.Content, "?") {
			s.ChannelTyping(m.ChannelID) // Send a typing event
			
			fmt.Println("### [START] Wolfram")
			
			query := m.Content
			fmt.Println("Original query: " + query)
			
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
			fmt.Println("Sanitized query: " + query)
			
			queryResultObject, err := wolframClient.GetQueryResult(query, nil)
			if err != nil {
				//s.ChannelMessageSend(m.ChannelID, botName + " was unable to process your request.\n" + fmt.Sprintf("%v", err))
				s.ChannelMessageSend(m.ChannelID, botName + " was unable to process your request.")
				fmt.Println(fmt.Sprintf("Error getting query result: %v", err))
				return
			}
			
			queryResult := queryResultObject.QueryResult
			pods := queryResult.Pods
			
			if len(pods) < 1 {
				s.ChannelMessageSend(m.ChannelID, botName + " was unable to process your request.")
				fmt.Println("Error getting pods from query")
				return
			}
			
			result := ""
			
			for _, pod := range pods {
				podTitle := pod.Title
				switch podTitle {
					case "Locations":
						fmt.Println("Denied pod: " + podTitle)
						continue
					case "Local map":
						fmt.Println("Denied pod: " + podTitle)
						continue
					case "Inferred local map":
						fmt.Println("Denied pod: " + podTitle)
						continue
					case "Inferred nearest city center":
						fmt.Println("Denied pod: " + podTitle)
						continue
					case "IP address":
						fmt.Println("Denied pod: " + podTitle)
						continue
					case "IP address registrant":
						fmt.Println("Denied pod: " + podTitle)
						continue
				}
				
				subPods := pod.SubPods
				if len(subPods) > 0 {
					for _, subPod := range subPods {
						plaintext := subPod.Plaintext
						if plaintext != "" {
							fmt.Println("Found result from pod [" + podTitle + "]: " + plaintext)
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
				s.ChannelMessageSend(m.ChannelID, botName + " was either unable to process your request or was denied permission from doing so.")
				fmt.Println("Error getting legal data from available pods")
				return
			}
			
			// Make nicer for Discord
			result = strings.Replace(result, "Wolfram|Alpha", botName, -1)
			result = strings.Replace(result, "Wolfram Alpha", botName, -1)
			result = strings.Replace(result, "I was created by Stephen Wolfram and his team.", "I was created by JoshuaDoes.", -1)
			
			s.ChannelMessageSend(m.ChannelID, result)
			
			fmt.Println("### [END]")
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
			fmt.Println("A> Leaving voice channel [" + guildID + ":" + channelID + "]...")
			playbackStopped[i] = true
			voiceConnectionRow.Disconnect()
			
			clearVoiceSession(i)
			
			return
		}
	}
}

func stopSound(guildID, channelID string) {
	for i, voiceConnectionRow := range voiceConnections {
		if voiceConnectionRow.ChannelID == channelID {
			fmt.Println("A> Stopping sound on voice channel [" + guildID + ":" + channelID + "]...")
			playbackStopped[i] = true
			return
		}
	}
}

func playSound(s *discordgo.Session, guildID, channelID string, callerChannelID string, url string) (err error) {
	var voiceConnection *discordgo.VoiceConnection = nil
	var encodingSession *dca.EncodeSession = nil
	var stream *dca.StreamingSession = nil
	var index int = -1
	for i, voiceConnectionRow := range voiceConnections {
		if voiceConnectionRow.ChannelID == channelID {
			fmt.Println("A> Found previous connection to voice channel [" + guildID + ":" + channelID + "]")
			voiceConnection = voiceConnections[i]
			encodingSession = encodingSessions[i]
			stream = streams[i]
			playbackStopped[i] = true
			index = i
			break
		}
	}
	
	time.Sleep(1000 * time.Millisecond)

	if voiceConnection == nil {
		fmt.Println("1B> Connecting to voice channel [" + guildID + ":" + channelID + "]...")
		voiceConnection, err := s.ChannelVoiceJoin(guildID, channelID, false, false)
		if err != nil {
			fmt.Println("1C> Error connecting to voice channel [" + guildID + ":" + channelID + "]")
			return err
		}
		
		fmt.Println("1D> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
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
				fmt.Println("1E> Error getting video info from [" + url + "]")
				return err
			}
			
			fmt.Println("1F> Storing video metadata...")
			title = videoInfo.Title
			author = videoInfo.Author
			//imageURL = videoInfo.GetThumbnailURL("maxresdefault").String()
			thumbnailURL = videoInfo.GetThumbnailURL("default").String()
			
			format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
			downloadURL, err := videoInfo.GetDownloadURL(format)
			if err != nil {
				fmt.Println("1G> Error getting download URL from [" + url + "]")
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
		
		encodingSession, err := dca.EncodeFile(mediaURL, options)
		if err != nil {
			fmt.Println("1I> Error encoding file [" + mediaURL + "]")
			return err
		}

		fmt.Println("1K> Setting speaking to true in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(true)

		done := make(chan error)
		stream := dca.NewStream(encodingSession, voiceConnection, done)
		
		fmt.Println("1L> Storing voiceConnection, encodingSession, stream, and playbackStopped handles/states in memory...")
		voiceConnections = append(voiceConnections, voiceConnection)
		encodingSessions = append(encodingSessions, encodingSession)
		streams = append(streams, stream)
		playbackStopped = append(playbackStopped, false)
		index = len(playbackStopped) - 1
		
		ticker := time.NewTicker(time.Second)
		
		for {
			if playbackStopped[index] == true {
				ticker.Stop()
				fmt.Println("1Q> Stopping encoding session...")
				encodingSession.Stop()
				fmt.Println("1R> Cleaning up encoding session...")
				encodingSession.Cleanup()
				fmt.Println("1S> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
				voiceConnection.Speaking(false)
				ticker.Stop()
				return nil
			}
			select {
				case err := <- done:
					if err != nil && err != io.EOF {
						fmt.Println("1M> Error creating stream")
						fmt.Println("1N> Cleaning up encoding session...")
						encodingSession.Stop()
						encodingSession.Cleanup()
						encodingSession.Truncate()
						fmt.Println("1O> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
						voiceConnection.Speaking(false)
						ticker.Stop()
						return err
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
		
		fmt.Println("1T> Cleaning up encoding session...")
		encodingSession.Stop()
		encodingSession.Cleanup()
		encodingSession.Truncate()

		fmt.Println("1U> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(false)
		
		ticker.Stop()

		return nil
	} else {
		fmt.Println("2B> Pausing stream...")
		stream.SetPaused(true)
		
		fmt.Println("2C> Cleaning up encoding session...")
		encodingSession.Stop()
		encodingSession.Cleanup()
		encodingSession.Truncate()

		fmt.Println("2D> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
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
				fmt.Println("1E> Error getting video info from [" + url + "]")
				return err
			}
			
			fmt.Println("1F> Storing video metadata...")
			title = videoInfo.Title
			author = videoInfo.Author
			//imageURL = videoInfo.GetThumbnailURL("maxresdefault").String()
			thumbnailURL = videoInfo.GetThumbnailURL("default").String()
			
			format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
			downloadURL, err := videoInfo.GetDownloadURL(format)
			if err != nil {
				fmt.Println("1G> Error getting download URL from [" + url + "]")
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
		
		encodingSession, err := dca.EncodeFile(mediaURL, options)
		if err != nil {
			fmt.Println("1I> Error encoding file [" + mediaURL + "]")
			return err
		}

		fmt.Println("1K> Setting speaking to true in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(true)

		done := make(chan error)
		stream := dca.NewStream(encodingSession, voiceConnection, done)
		
		ticker := time.NewTicker(time.Second)
		
		for {
			if playbackStopped[index] == true {
				ticker.Stop()
				fmt.Println("1Q> Stopping encoding session...")
				encodingSession.Stop()
				fmt.Println("1R> Cleaning up encoding session...")
				encodingSession.Cleanup()
				fmt.Println("1S> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
				voiceConnection.Speaking(false)
				ticker.Stop()
				return nil
			}
			select {
				case err := <- done:
					if err != nil && err != io.EOF {
						fmt.Println("1M> Error creating stream")
						fmt.Println("1N> Cleaning up encoding session...")
						encodingSession.Stop()
						encodingSession.Cleanup()
						encodingSession.Truncate()
						fmt.Println("1O> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
						voiceConnection.Speaking(false)
						ticker.Stop()
						return err
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
		
		fmt.Println("1T> Cleaning up encoding session...")
		encodingSession.Stop()
		encodingSession.Cleanup()
		encodingSession.Truncate()

		fmt.Println("1U> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(false)
		
		ticker.Stop()

		return nil
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