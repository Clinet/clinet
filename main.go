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

	"github.com/paked/configure"
	"github.com/bwmarrin/discordgo"
	"github.com/Krognol/go-wolfram"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
)

var (
	conf = configure.New()
	confBotToken = conf.String("botToken", "", "Bot Token")
	confBotName = conf.String("botName", "", "Bot Name")
	confBotPrefix = conf.String("botPrefix", "", "Bot Prefix")
	confWolframAppID = conf.String("wolframAppID", "", "Wolfram AppID")
	botToken string = ""
	botName string = ""
	botPrefix string = ""
	wolframAppID string = ""

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
	if (botToken == "" || botName == "" || botPrefix == "" || wolframAppID == "") {
		fmt.Println("> Configuration not properly setup, exiting...")
		return
	} else {
		fmt.Println("> Successfully loaded configuration.")
		fmt.Println("botToken: " + botToken)
		fmt.Println("botName: " + botName)
		fmt.Println("botPrefix: " + botPrefix)
		fmt.Println("wolframAppID: " + wolframAppID)
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
		if strings.HasPrefix(m.Content, botPrefix + "play ") {
			url := strings.Replace(m.Content, botPrefix + "play ", "", -1)
			if url == "" {
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
				}
			}
		}
	} else if strings.Contains(m.Content, botName) {
		if strings.HasSuffix(m.Content, "?") {
			query := m.Content
			
			//Sanitize for Wolfram|Alpha
			query = strings.Replace(query, botName, "", -1)
			query = strings.Replace(query, ",", "", -1)
			
			result, err := wolframClient.GetShortAnswerQuery(query, 0, 0)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Error processing request.")
				return
			}
			fmt.Println("Wolfram: " + result)
			s.ChannelMessageSend(m.ChannelID, result)
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
		
		videoInfo, err := ytdl.GetVideoInfo(url)
		if err != nil {
			fmt.Println("1E> Error getting video info from [" + url + "]")
			return err
		}
		
		fmt.Println("1F> Storing video metadata...")
		title := videoInfo.Title
		author := videoInfo.Author
		imageURL := videoInfo.GetThumbnailURL("maxresdefault").String()
		thumbnailURL := videoInfo.GetThumbnailURL("default").String()
		
		format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
		downloadURL, err := videoInfo.GetDownloadURL(format)
		if err != nil {
			fmt.Println("1G> Error getting download URL from [" + url + "]")
			return err
		}
		
		encodingSession, err := dca.EncodeFile(downloadURL.String(), options)
		if err != nil {
			fmt.Println("1I> Error encoding file [" + downloadURL.String() + "]")
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
		
		embed := NewEmbed().
			SetTitle(title).
			SetDescription(author).
			SetImage(imageURL).
			SetThumbnail(thumbnailURL).
			SetColor(0xff0000).MessageEmbed
		s.ChannelMessageSendEmbed(callerChannelID, embed)
		
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
						return err
					}
				case <- ticker.C:
					stats := encodingSession.Stats()
					playbackPosition := stream.PlaybackPosition()

					fmt.Printf("Playback: %10s, Transcode Stats: Time: %5s, Size: %5dkB, Bitrate: %6.2fkB, Speed: %5.1fx\r", playbackPosition, stats.Duration.String(), stats.Size, stats.Bitrate, stats.Speed)
			}
		}
		
		fmt.Println("1T> Cleaning up encoding session...")
		encodingSession.Stop()
		encodingSession.Cleanup()
		encodingSession.Truncate()

		fmt.Println("1U> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(false)

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
		
		videoInfo, err := ytdl.GetVideoInfo(url)
		if err != nil {
			fmt.Println("2E> Error getting video info from [" + url + "]")
			return err
		}
		
		fmt.Println("2F> Storing video metadata...")
		title := videoInfo.Title
		author := videoInfo.Author
		imageURL := videoInfo.GetThumbnailURL("maxresdefault").String()
		thumbnailURL := videoInfo.GetThumbnailURL("default").String()
		
		format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
		downloadURL, err := videoInfo.GetDownloadURL(format)
		if err != nil {
			fmt.Println("2G> Error getting download URL from [" + url + "]")
			return err
		}
		
		encodingSession, err := dca.EncodeFile(downloadURL.String(), options)
		if err != nil {
			fmt.Println("2I> Error encoding file [" + downloadURL.String() + "]")
			return err
		}
		
		fmt.Println("2K> Setting speaking to true in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(true)

		done := make(chan error)
		stream := dca.NewStream(encodingSession, voiceConnection, done)
		
		embed := NewEmbed().
			SetTitle(title).
			SetDescription(author).
			SetImage(imageURL).
			SetThumbnail(thumbnailURL).
			SetColor(0xff0000).MessageEmbed
		s.ChannelMessageSendEmbed(callerChannelID, embed)
		
		fmt.Println("2L> Setting playbackStopped to false...")
		playbackStopped[index] = false
		
		ticker := time.NewTicker(time.Second)
		
		for {
			if playbackStopped[index] == true {
				ticker.Stop()
				fmt.Println("2Q> Stopping encoding session...")
				encodingSession.Stop()
				fmt.Println("2R> Cleaning up encoding session...")
				encodingSession.Cleanup()
				fmt.Println("2S> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
				voiceConnection.Speaking(false)
				return nil
			}
			select {
				case err := <- done:
					if err != nil && err != io.EOF {
						fmt.Println("2M> Error creating stream")
						fmt.Println("2N> Cleaning up encoding session...")
						encodingSession.Stop()
						encodingSession.Cleanup()
						encodingSession.Truncate()
						fmt.Println("2O> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
						voiceConnection.Speaking(false)
						return err
					}
				case <- ticker.C:
					stats := encodingSession.Stats()
					playbackPosition := stream.PlaybackPosition()

					fmt.Printf("Playback: %10s, Transcode Stats: Time: %5s, Size: %5dkB, Bitrate: %6.2fkB, Speed: %5.1fx\r", playbackPosition, stats.Duration.String(), stats.Size, stats.Bitrate, stats.Speed)
			}
		}
		
		fmt.Println("2T> Cleaning up encoding session...")
		encodingSession.Stop()
		encodingSession.Cleanup()
		encodingSession.Truncate()

		fmt.Println("2U> Setting speaking to false in voice channel [" + voiceConnection.GuildID + ":" + voiceConnection.ChannelID + "]...")
		voiceConnection.Speaking(false)

		return nil
	}
}