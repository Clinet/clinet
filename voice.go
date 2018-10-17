package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"

	"github.com/JoshuaDoes/goprobe"
	"github.com/JoshuaDoes/spotigo"
	"github.com/bwmarrin/discordgo"
	isoduration "github.com/channelmeter/iso8601duration"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
	youtube "google.golang.org/api/youtube/v3"
)

var encodeOptionsPresetHigh = &dca.EncodeOptions{
	Volume:           256,
	Channels:         2,
	FrameRate:        48000,
	FrameDuration:    20,
	Bitrate:          128,
	Application:      "lowdelay",
	CompressionLevel: 10,
	PacketLoss:       0,
	BufferedFrames:   100,
	VBR:              true,
	RawOutput:        true,
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

type SpotifyResultNav struct {
	Query        string
	TotalResults int
	AllResults   []spotigo.SpotigoSearchHit //All results
	Results      []spotigo.SpotigoSearchHit //Results for current page

	PlaylistID     string
	PlaylistUserID string
	IsPlaylist     bool

	PageNumber int
	MaxResults int
}

func (page *SpotifyResultNav) GetResults() ([]spotigo.SpotigoSearchHit, error) {
	if len(page.Results) == 0 {
		return nil, errors.New("No search results found")
	}
	return page.Results, nil
}
func (page *SpotifyResultNav) Search(query string) error {
	page.Query = ""
	page.TotalResults = 0
	page.Results = nil
	page.IsPlaylist = false

	trackResults, err := botData.BotClients.Spotify.SearchTracks(query)
	if err != nil {
		return err
	}

	page.Query = query
	page.TotalResults = len(trackResults.Hits)
	page.Results = trackResults.Hits

	return nil
}
func (page *SpotifyResultNav) Playlist(url string) error {
	if page.MaxResults == 0 {
		page.MaxResults = botData.BotOptions.SpotifyMaxResults
	}

	page.Query = ""
	page.TotalResults = 0
	page.Results = nil
	page.PageNumber = 0
	page.IsPlaylist = false

	playlist, err := botData.BotClients.Spotify.GetPlaylist(url)
	if err != nil {
		return err
	}

	playlistItems := make([]spotigo.SpotigoSearchHit, 0)
	for i := 0; i < len(playlist.Contents.Items); i++ {
		item := playlist.Contents.Items[i]
		hit := spotigo.SpotigoSearchHit{}

		trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(item.TrackURI)
		if err != nil {
			continue
		}

		artists := make([]spotigo.SpotigoSearchHitArtist, 0)
		artists = append(artists, spotigo.SpotigoSearchHitArtist{Name: trackInfo.Artist})

		hit.URI = item.TrackURI
		hit.ID = trackInfo.TrackID
		hit.Name = trackInfo.Title
		hit.ImageURL = trackInfo.ArtURL
		hit.Duration = trackInfo.Duration
		hit.Artists = artists

		playlistItems = append(playlistItems, hit)
	}

	page.PageNumber = 1
	page.IsPlaylist = true
	page.Query = playlist.Attributes.Name
	page.AllResults = playlistItems
	page.Results = page.AllResults[(page.PageNumber-1)*page.MaxResults : page.PageNumber*page.MaxResults]
	page.TotalResults = len(page.AllResults)
	page.PlaylistID = playlist.PlaylistID
	page.PlaylistUserID = playlist.UserID

	return nil
}
func (page *SpotifyResultNav) Prev() error {
	if page.PageNumber == 0 {
		return errors.New("No pages found")
	}
	if ((page.PageNumber-1)*page.MaxResults) > page.TotalResults || ((page.PageNumber-1)*page.MaxResults) <= 0 {
		return errors.New("Page not available")
	}

	page.PageNumber--
	page.Results = page.AllResults[(page.PageNumber-1)*page.MaxResults : page.PageNumber*page.MaxResults]

	return nil
}
func (page *SpotifyResultNav) Next() error {
	if page.PageNumber == 0 {
		return errors.New("No pages found")
	}
	if ((page.PageNumber+1)*page.MaxResults) > page.TotalResults || ((page.PageNumber+1)*page.MaxResults) <= 0 {
		return errors.New("Page not available")
	}

	page.PageNumber++
	page.Results = page.AllResults[(page.PageNumber-1)*page.MaxResults : page.PageNumber*page.MaxResults]

	return nil
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
		MaxResults(page.MaxResults).
		PageToken(page.PrevPageToken)

	response, err := searchCall.Do()
	if err != nil {
		return errors.New("Could not find any video results for the previous page")
	}

	page.PageNumber--
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
		MaxResults(page.MaxResults).
		PageToken(page.NextPageToken)

	response, err := searchCall.Do()
	if err != nil {
		return errors.New("Could not find any video results for the next page")
	}

	page.PageNumber++
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
		page.MaxResults = int64(botData.BotOptions.YouTubeMaxResults)
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

type AudioQueueEntry struct {
	Author           string
	ImageURL         string
	MediaURL         string
	Requester        *discordgo.User
	RequestMessageID string //Used for if someone edits their request
	ThumbnailURL     string
	Title            string
	Type             string
	Duration         float64
	TrackID          string //Used for instances like Spotify where the media URL is not allowed to be directly served
}

func (audioQueueEntry *AudioQueueEntry) GetNowPlayingEmbed() *discordgo.MessageEmbed {
	switch audioQueueEntry.Type {
	case "youtube":
		return NewEmbed().
			AddField("Now Playing from YouTube", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF0000).MessageEmbed
	case "soundcloud":
		return NewEmbed().
			AddField("Now Playing from SoundCloud", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF7700).MessageEmbed
	case "spotify":
		return NewEmbed().
			AddField("Now Playing from Spotify", "["+audioQueueEntry.Title+"](https://open.spotify.com/track/"+audioQueueEntry.TrackID+") by **"+audioQueueEntry.Author+"**").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0x1DB954).MessageEmbed
	default:
		if audioQueueEntry.Author != "" && audioQueueEntry.Title != "" {
			return NewEmbed().
				AddField("Now Playing", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
				AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
				AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
				SetColor(0x1C1C1C).MessageEmbed
		}
		return NewEmbed().
			SetTitle("Now Playing").
			AddField("URL", audioQueueEntry.MediaURL).
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetColor(0x1C1C1C).MessageEmbed
	}
}
func (audioQueueEntry *AudioQueueEntry) GetNowPlayingDurationEmbed(stream *dca.StreamingSession) *discordgo.MessageEmbed {
	//Get the current duration
	playbackPosition := secondsToHuman(stream.PlaybackPosition().Seconds())
	fullDuration := secondsToHuman(audioQueueEntry.Duration)

	switch audioQueueEntry.Type {
	case "youtube":
		return NewEmbed().
			AddField("Now Playing from YouTube", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Time", playbackPosition+" / "+fullDuration).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF0000).MessageEmbed
	case "soundcloud":
		return NewEmbed().
			AddField("Now Playing from SoundCloud", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Time", playbackPosition+" / "+fullDuration).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF7700).MessageEmbed
	case "spotify":
		return NewEmbed().
			AddField("Now Playing from Spotify", "["+audioQueueEntry.Title+"](https://open.spotify.com/track/"+audioQueueEntry.TrackID+") by **"+audioQueueEntry.Author+"**").
			AddField("Time", playbackPosition+" / "+fullDuration).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0x1DB954).MessageEmbed
	default:
		if audioQueueEntry.Author != "" && audioQueueEntry.Title != "" {
			return NewEmbed().
				AddField("Now Playing", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
				AddField("Time", playbackPosition+" / "+fullDuration).
				AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
				SetColor(0x1C1C1C).MessageEmbed
		}
		return NewEmbed().
			SetTitle("Now Playing").
			AddField("URL", audioQueueEntry.MediaURL).
			AddField("Time", playbackPosition+" / "+fullDuration).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetColor(0x1C1C1C).MessageEmbed
	}
}
func (audioQueueEntry *AudioQueueEntry) GetQueueAddedEmbed() *discordgo.MessageEmbed {
	if audioQueueEntry.Type == "" {
		if isYouTubeURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "youtube"
		} else if isSoundCloudURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "soundcloud"
		} else if isSpotifyTrackURLURI(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "spotify"
		} else {
			audioQueueEntry.Type = "direct"
		}
	}

	switch audioQueueEntry.Type {
	case "youtube":
		return NewEmbed().
			AddField("Added to Queue from YouTube", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF0000).MessageEmbed
	case "soundcloud":
		return NewEmbed().
			AddField("Added to Queue from SoundCloud", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0xFF7700).MessageEmbed
	case "spotify":
		return NewEmbed().
			AddField("Added to Queue from Spotify", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetThumbnail(audioQueueEntry.ThumbnailURL).
			SetColor(0x1DB954).MessageEmbed
	default:
		if audioQueueEntry.Author != "" && audioQueueEntry.Title != "" {
			return NewEmbed().
				AddField("Added to Queue", "["+audioQueueEntry.Title+"]("+audioQueueEntry.MediaURL+") by **"+audioQueueEntry.Author+"**").
				AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
				AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
				SetColor(0x1C1C1C).MessageEmbed
		}
		return NewEmbed().
			SetTitle("Added to Queue").
			AddField("Duration", secondsToHuman(audioQueueEntry.Duration)).
			AddField("URL", audioQueueEntry.MediaURL).
			AddField("Requester", "<@"+audioQueueEntry.Requester.ID+">").
			SetColor(0x1C1C1C).MessageEmbed
	}
}
func (audioQueueEntry *AudioQueueEntry) FillMetadata() *discordgo.MessageEmbed {
	if audioQueueEntry.Type == "" {
		if isYouTubeURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "youtube"
		} else if isSoundCloudURL(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "soundcloud"
		} else if isSpotifyTrackURLURI(audioQueueEntry.MediaURL) {
			audioQueueEntry.Type = "spotify"
		} else {
			audioQueueEntry.Type = "direct"
		}
	}

	probeURL, err := getMediaURL(audioQueueEntry.MediaURL)
	if err != nil {
		return NewErrorEmbed("Voice Error", "Error probing media for audio: Could not find the media URL. Err: "+fmt.Sprintf("%v", err))
	}

	probe, err := goprobe.ProbeMedia(probeURL)
	if err != nil {
		return NewErrorEmbed("Voice Error", "Error probing the media for audio.")
	}
	if len(probe.Streams) == 0 {
		return NewErrorEmbed("Voice Error", "Error probing the media for audio: No streams were found.")
	}

	hasAudio := false
	audioStream := 0
	for i, stream := range probe.Streams {
		if stream.CodecType == "audio" {
			hasAudio = true
			audioStream = i
			break
		}
	}
	if !hasAudio {
		return NewErrorEmbed("Voice Error", "Error probing the media for audio: Media does not have audio.")
	}

	switch audioQueueEntry.Type {
	case "youtube":
		videoInfo, err := ytdl.GetVideoInfo(audioQueueEntry.MediaURL)
		if err != nil {
			return NewErrorEmbed("Voice Error", "Error getting info about the YouTube video.")
		}
		audioQueueEntry.Author = videoInfo.Author
		audioQueueEntry.ImageURL = videoInfo.GetThumbnailURL("maxresdefault").String()
		audioQueueEntry.ThumbnailURL = videoInfo.GetThumbnailURL("default").String()
		audioQueueEntry.Title = videoInfo.Title

		call := youtube.NewVideosService(botData.BotClients.YouTube).
			List("contentDetails").
			Id(videoInfo.ID)

		response, err := call.Do()
		if err != nil {
			return nil
		}

		duration, err := isoduration.FromString(response.Items[0].ContentDetails.Duration)
		if err != nil {
			return nil
		}
		audioQueueEntry.Duration = duration.ToDuration().Seconds()
	case "soundcloud":
		audioInfo, err := botData.BotClients.SoundCloud.GetTrackInfo(audioQueueEntry.MediaURL)
		if err != nil {
			return NewErrorEmbed("Voice Error", "Error getting info about the SoundCloud track.")
		}
		audioQueueEntry.Author = audioInfo.Artist
		audioQueueEntry.ImageURL = audioInfo.ArtURL
		audioQueueEntry.ThumbnailURL = audioInfo.ArtURL
		audioQueueEntry.Title = audioInfo.Title
		audioQueueEntry.Duration, _ = strconv.ParseFloat(probe.Streams[audioStream].Duration, 64)
	case "spotify":
		audioInfo, err := botData.BotClients.Spotify.GetTrackInfo(audioQueueEntry.MediaURL)
		if err != nil {
			return NewErrorEmbed("Voice Error", "Error getting info about the Spotify track.")
		}
		audioQueueEntry.Author = audioInfo.Artist
		audioQueueEntry.ImageURL = audioInfo.ArtURL
		audioQueueEntry.ThumbnailURL = audioInfo.ArtURL
		audioQueueEntry.Title = audioInfo.Title
		audioQueueEntry.Duration = float64(audioInfo.Duration / 1000)
		audioQueueEntry.TrackID = audioInfo.TrackID
	default:
		audioQueueEntry.Author = probe.Format.Tags.Artist
		audioQueueEntry.Title = probe.Format.Tags.Title
		audioQueueEntry.Duration, _ = strconv.ParseFloat(probe.Streams[audioStream].Duration, 64)
	}

	return nil
}

// VoiceData contains data about the current voice session
type VoiceData struct {
	VoiceConnection  *discordgo.VoiceConnection `json:"-"` //Holds data regarding the current voice connection
	EncodingSession  *dca.EncodeSession         `json:"-"` //Holds data regarding the current encoding session
	EncodingOptions  *dca.EncodeOptions         //Holds data regarding the current encoding options
	StreamingSession *dca.StreamingSession      `json:"-"` //Holds data regarding the current streaming session

	ChannelIDJoinedFrom string //The text channel that was used to bring the bot into the voice channel

	IsPlaybackPreparing bool `json:"-"` //Whether or not the playback is being prepared
	IsPlaybackRunning   bool `json:"-"` //Whether or not playback is currently running
	WasStoppedManually  bool `json:"-"` //Whether or not playback was stopped manually or automatically
	WasSkipped          bool `json:"-"` //Whether or not playback was skipped

	//Configuration settings that can be set via commands
	RepeatLevel int //0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing
	Shuffle     bool
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
			guildData[guildID].VoiceData = VoiceData{}
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
	if guildData[guildID].VoiceData.EncodingOptions == nil {
		guildData[guildID].VoiceData.EncodingOptions = encodeOptionsPresetHigh
	}
	guildData[guildID].VoiceData.EncodingOptions.RawOutput = true
	guildData[guildID].VoiceData.EncodingOptions.Bitrate = 96
	guildData[guildID].VoiceData.EncodingOptions.Application = "lowdelay"

	//Create the encoding session to encode the audio to DCA in a stream
	guildData[guildID].VoiceData.EncodingSession, err = dca.EncodeFile(mediaURL, guildData[guildID].VoiceData.EncodingOptions)
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

	//Set speaking to false
	guildData[guildID].VoiceData.VoiceConnection.Speaking(false)

	//Check streaming session for why playback stopped
	if guildData[guildID].VoiceData.StreamingSession != nil {
		_, err = guildData[guildID].VoiceData.StreamingSession.Finished()
	}

	//Clean up streaming session
	guildData[guildID].VoiceData.StreamingSession = nil

	//Clean up encoding session
	if guildData[guildID].VoiceData.EncodingSession != nil {
		guildData[guildID].VoiceData.EncodingSession.Stop()
		guildData[guildID].VoiceData.EncodingSession.Cleanup()
		guildData[guildID].VoiceData.EncodingSession = nil
	}

	//If playback stopped from an error, return that error
	if err != nil {
		return err
	}
	return nil
}

func voicePlayWrapper(session *discordgo.Session, guildID, channelID, mediaURL string) {
	err := voicePlay(guildID, mediaURL)
	if guildData[guildID].VoiceData.RepeatLevel == 2 { //Repeat Now Playing
		for guildData[guildID].VoiceData.RepeatLevel == 2 {
			err = voicePlay(guildID, mediaURL)
			if err != nil {
				guildData[guildID].AudioNowPlaying = AudioQueueEntry{} //Clear now playing slot
				errorEmbed := NewErrorEmbed("Voice Error", "There was an error playing the specified audio.")
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
		errorEmbed := NewErrorEmbed("Voice Error", "There was an error playing the specified audio.")
		session.ChannelMessageSendEmbed(channelID, errorEmbed)
		return
	} else {
		if guildData[guildID].VoiceData.WasStoppedManually {
			guildData[guildID].VoiceData.WasStoppedManually = false
		} else if guildData[guildID].VoiceData.IsPlaybackRunning == false || guildData[guildID].VoiceData.WasSkipped == true {
			guildData[guildID].VoiceData.WasSkipped = false //Reset skip bool in case it was true

			//When the song finishes playing, we should run on a loop to make sure the next songs continue playing
			for len(guildData[guildID].AudioQueue) > 0 {
				if guildData[guildID].VoiceData.WasStoppedManually {
					guildData[guildID].VoiceData.WasStoppedManually = false
					return //Prevent next guild queue entry from playing
				}

				//Move next guild queue entry into now playing slot
				if guildData[guildID].VoiceData.Shuffle {
					//Pseudo-shuffle™, replace with legitimate shuffle method later so user can return to non-shuffled playlist
					randomEntry := rand.Intn(len(guildData[guildID].AudioQueue))
					guildData[guildID].AudioNowPlaying = guildData[guildID].AudioQueue[randomEntry]
					guildData[guildID].QueueRemove(randomEntry)
				} else {
					guildData[guildID].AudioNowPlaying = guildData[guildID].AudioQueue[0]
					guildData[guildID].QueueRemove(0)
				}

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
							errorEmbed := NewErrorEmbed("Voice Error", "There was an error playing the specified audio.")
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
					errorEmbed := NewErrorEmbed("Voice Error", "There was an error playing the specified audio.")
					session.ChannelMessageSendEmbed(channelID, errorEmbed)
					return //Prevent next guild queue entry from playing
				} else {
					if guildData[guildID].VoiceData.WasStoppedManually {
						guildData[guildID].VoiceData.WasStoppedManually = false
						return //Prevent next guild queue entry from playing
					}
				}
			}

			if guildData[guildID].VoiceData.WasStoppedManually == false {
				voiceLeave(guildID, channelID) //We're done with everything so leave the voice channel
				leaveEmbed := NewGenericEmbed("Voice", "Finished playing the queue.")
				session.ChannelMessageSendEmbed(channelID, leaveEmbed)
			}
		}
	}
}

func voiceStop(guildID string) {
	if guildData[guildID] != nil {
		guildData[guildID].VoiceData.WasStoppedManually = true //Make sure other threads know it was stopped manually
		if guildData[guildID].VoiceData.EncodingSession != nil {
			guildData[guildID].VoiceData.EncodingSession.Stop()    //Stop the encoding session manually
			guildData[guildID].VoiceData.EncodingSession.Cleanup() //Cleanup the encoding session
		}
		guildData[guildID].VoiceData.IsPlaybackRunning = false //Let the voice play function clean up on its own
	}
}

func voiceSkip(guildID string) {
	if guildData[guildID] != nil {
		guildData[guildID].VoiceData.WasSkipped = true //Let the voice play wrapper function continue to the next song if available
		if guildData[guildID].VoiceData.EncodingSession != nil {
			guildData[guildID].VoiceData.EncodingSession.Stop()    //Stop the encoding session manually
			guildData[guildID].VoiceData.EncodingSession.Cleanup() //Cleanup the encoding session
		}
		guildData[guildID].VoiceData.IsPlaybackRunning = false //Let the voice play function clean up on its own
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

	if isSpotifyTrackURLURI(url) {
		audioInfo, err := botData.BotClients.Spotify.GetTrackInfo(url)
		if err != nil {
			return url, err
		}

		return audioInfo.StreamURL, nil
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
func isSpotifyTrackURLURI(url string) bool {
	regexpHasSpotify, _ := regexp.MatchString("^(https:\\/\\/open.spotify.com\\/track\\/|spotify:track:)([a-zA-Z0-9]+)(.*)$", url)
	if regexpHasSpotify {
		return true
	}
	return false
}
func isSpotifyPlaylistURLURI(url string) bool {
	regexHasSpotify, _ := regexp.MatchString("^(https:\\/\\/open.spotify.com\\/user\\/|spotify:user:)([a-zA-Z0-9]+)(\\/playlist\\/|:playlist:)([a-zA-Z0-9]+)(.*)$", url)
	if regexHasSpotify {
		return true
	}
	return false
}
