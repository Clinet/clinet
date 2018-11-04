package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"github.com/JoshuaDoes/spotigo"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	youtube "google.golang.org/api/youtube/v3"
)

var encodeOptionsPresetHigh = &dca.EncodeOptions{
	Volume:           256,
	Channels:         2,
	FrameRate:        48000,
	FrameDuration:    20,
	Bitrate:          128,
	Application:      "audio",
	CompressionLevel: 10,
	PacketLoss:       0,
	BufferedFrames:   100,
	VBR:              true,
	RawOutput:        true,
}

func (guild *GuildData) QueueAdd(queueEntry *QueueEntry) {
	guild.AudioQueue = append(guild.AudioQueue, queueEntry)
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
func (guild *GuildData) QueueGet(guildID string, entry int) *QueueEntry {
	if len(guildData[guildID].AudioQueue) >= entry {
		return guildData[guildID].AudioQueue[entry]
	} else {
		return nil
	}
}
func (guild *GuildData) QueueGetNext(guildID string) *QueueEntry {
	if len(guildData[guildID].AudioQueue) > 0 {
		return guildData[guildID].AudioQueue[0]
	} else {
		return nil
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
	TotalPages int

	AddingAll  bool
	AddedSoFar int
	Cancelled  bool

	GuildID string //To know what guild this page belongs to
}

func (page *SpotifyResultNav) GetResults() ([]spotigo.SpotigoSearchHit, error) {
	if len(page.Results) == 0 {
		return nil, errors.New("No search results found")
	}
	return page.Results, nil
}
func (page *SpotifyResultNav) Search(query string) error {
	if page.MaxResults == 0 {
		page.MaxResults = botData.BotOptions.SpotifyMaxResults
	}

	page.Query = ""
	page.TotalResults = 0
	page.AllResults = nil
	page.Results = nil
	page.PageNumber = 0
	page.IsPlaylist = false
	page.PlaylistID = ""
	page.PlaylistUserID = ""
	page.TotalPages = 0

	searchResults, err := botData.BotClients.Spotify.Search(query)
	if err != nil {
		return err
	}

	page.PageNumber = 1
	page.Query = query

	if len(searchResults.Results.Artists.Hits) > 0 {
		page.AllResults = append(page.AllResults, searchResults.Results.Artists.Hits...)
	}
	if len(searchResults.Results.Tracks.Hits) > 0 {
		page.AllResults = append(page.AllResults, searchResults.Results.Tracks.Hits...)
	}
	if len(searchResults.Results.Albums.Hits) > 0 {
		page.AllResults = append(page.AllResults, searchResults.Results.Albums.Hits...)
	}
	if len(searchResults.Results.Playlists.Hits) > 0 {
		page.AllResults = append(page.AllResults, searchResults.Results.Playlists.Hits...)
	}

	maxResults := page.MaxResults
	if len(page.AllResults) < page.MaxResults {
		maxResults = len(page.AllResults)
	}

	page.Results = page.AllResults[(page.PageNumber-1)*page.MaxResults : page.PageNumber*maxResults]
	page.TotalResults = len(page.AllResults)
	page.TotalPages = int(math.Ceil(float64(page.TotalResults) / float64(page.MaxResults)))

	return nil
}
func (page *SpotifyResultNav) Playlist(url string) error {
	if page.MaxResults == 0 {
		page.MaxResults = botData.BotOptions.SpotifyMaxResults
	}

	page.Query = ""
	page.TotalResults = 0
	page.AllResults = nil
	page.Results = nil
	page.PageNumber = 0
	page.IsPlaylist = false
	page.PlaylistID = ""
	page.PlaylistUserID = ""
	page.TotalPages = 0

	playlist, err := botData.BotClients.Spotify.GetPlaylist(url)
	if err != nil {
		return err
	}
	if len(playlist.Contents.Items) <= 0 {
		return errors.New("no tracks found")
	}

	playlistItems := make([]spotigo.SpotigoSearchHit, 0)
	for i := 0; i < len(playlist.Contents.Items); i++ {
		//Give a chance for other commands waiting in line to execute
		guildData[page.GuildID].Unlock()
		guildData[page.GuildID].Lock()

		if page.Cancelled {
			page.Cancelled = false
			break
		}

		item := playlist.Contents.Items[i]
		hit := spotigo.SpotigoSearchHit{}

		if i < page.MaxResults {
			trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(item.TrackURI)
			if err != nil {
				continue
			}

			artists := make([]spotigo.SpotigoSearchHitArtist, 0)
			artists = append(artists, spotigo.SpotigoSearchHitArtist{Name: trackInfo.Artist})

			hit.ID = trackInfo.TrackID
			hit.Name = trackInfo.Title
			hit.ImageURL = trackInfo.ArtURL
			hit.Duration = trackInfo.Duration
			hit.Artists = artists
		}

		hit.URI = item.TrackURI

		playlistItems = append(playlistItems, hit)
	}

	if len(playlistItems) <= 0 {
		return errors.New("no tracks found")
	}

	maxResults := page.MaxResults
	if len(playlistItems) < page.MaxResults {
		maxResults = len(playlistItems)
	}

	page.PageNumber = 1
	page.IsPlaylist = true
	page.Query = playlist.Attributes.Name
	page.AllResults = playlistItems
	page.Results = page.AllResults[(page.PageNumber-1)*page.MaxResults : page.PageNumber*maxResults]
	page.TotalResults = len(page.AllResults)
	page.PlaylistID = playlist.PlaylistID
	page.PlaylistUserID = playlist.UserID
	page.TotalPages = int(math.Ceil(float64(page.TotalResults) / float64(page.MaxResults)))

	return nil
}
func (page *SpotifyResultNav) Prev() error {
	if page.PageNumber == 0 {
		return errors.New("No pages found")
	}
	if (page.PageNumber - 1) <= 0 {
		return errors.New("Page not available")
	}

	page.PageNumber--
	low := (page.PageNumber - 1) * page.MaxResults
	high := page.PageNumber * page.MaxResults
	if high > len(page.AllResults) {
		high = len(page.AllResults)
	}
	page.Results = page.AllResults[low:high]

	for i := 0; i < len(page.Results); i++ {
		if strings.HasPrefix(page.Results[i].URI, "spotify:track:") {
			trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(page.Results[i].URI)
			if err != nil {
				continue
			}

			artists := make([]spotigo.SpotigoSearchHitArtist, 0)
			artists = append(artists, spotigo.SpotigoSearchHitArtist{Name: trackInfo.Artist})

			page.Results[i].ID = trackInfo.TrackID
			page.Results[i].Name = trackInfo.Title
			page.Results[i].ImageURL = trackInfo.ArtURL
			page.Results[i].Duration = trackInfo.Duration
			page.Results[i].Artists = artists
		}
	}

	return nil
}
func (page *SpotifyResultNav) Next() error {
	if page.PageNumber == 0 {
		return errors.New("No pages found")
	}
	if (page.PageNumber + 1) > page.TotalPages {
		return errors.New("Page not available")
	}

	page.PageNumber++
	low := (page.PageNumber - 1) * page.MaxResults
	high := page.PageNumber * page.MaxResults
	if high > len(page.AllResults) {
		high = len(page.AllResults)
	}
	page.Results = page.AllResults[low:high]

	for i := 0; i < len(page.Results); i++ {
		if strings.HasPrefix(page.Results[i].URI, "spotify:track:") {
			trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(page.Results[i].URI)
			if err != nil {
				continue
			}

			artists := make([]spotigo.SpotigoSearchHitArtist, 0)
			artists = append(artists, spotigo.SpotigoSearchHitArtist{Name: trackInfo.Artist})

			page.Results[i].ID = trackInfo.TrackID
			page.Results[i].Name = trackInfo.Title
			page.Results[i].ImageURL = trackInfo.ArtURL
			page.Results[i].Duration = trackInfo.Duration
			page.Results[i].Artists = artists
		}
	}

	return nil
}
func (page *SpotifyResultNav) Jump(pageNumber int) error {
	if page.PageNumber == 0 {
		return errors.New("No pages found")
	}
	if pageNumber > page.TotalPages || pageNumber < 1 {
		return errors.New("Page not available")
	}

	page.PageNumber = pageNumber
	low := (page.PageNumber - 1) * page.MaxResults
	high := page.PageNumber * page.MaxResults
	if high > len(page.AllResults) {
		high = len(page.AllResults)
	}
	page.Results = page.AllResults[low:high]

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

func (queueEntry *QueueEntry) GetNowPlayingEmbed() *discordgo.MessageEmbed {
	track := "[" + queueEntry.Metadata.Title + "](" + queueEntry.Metadata.DisplayURL + ")"
	if len(queueEntry.Metadata.Artists) > 0 {
		track += " by [" + queueEntry.Metadata.Artists[0].Name + "](" + queueEntry.Metadata.Artists[0].URL + ")"
		if len(queueEntry.Metadata.Artists) > 1 {
			track += " ft. " + "[" + queueEntry.Metadata.Artists[1].Name + "](" + queueEntry.Metadata.Artists[1].URL + ")"
			if len(queueEntry.Metadata.Artists) > 2 {
				for i, artist := range queueEntry.Metadata.Artists[2:] {
					track += ", "
					if (i + 3) == len(queueEntry.Metadata.Artists) {
						track += " and "
					}
					track += "[" + artist.Name + "](" + artist.URL + ")"
				}
			}
		}
	}
	return NewEmbed().
		AddField("Now Playing from "+queueEntry.ServiceName, track).
		AddField("Duration", secondsToHuman(queueEntry.Metadata.Duration)).
		AddField("Requester", "<@!"+queueEntry.Requester.ID+">").
		SetThumbnail(queueEntry.Metadata.ThumbnailURL).
		SetColor(queueEntry.ServiceColor).MessageEmbed
}
func (queueEntry *QueueEntry) GetNowPlayingDurationEmbed(stream *dca.StreamingSession) *discordgo.MessageEmbed {
	playbackPosition := secondsToHuman(stream.PlaybackPosition().Seconds())
	fullDuration := secondsToHuman(queueEntry.Metadata.Duration)

	track := "[" + queueEntry.Metadata.Title + "](" + queueEntry.Metadata.DisplayURL + ")"
	if len(queueEntry.Metadata.Artists) > 0 {
		track += " by [" + queueEntry.Metadata.Artists[0].Name + "](" + queueEntry.Metadata.Artists[0].URL + ")"
		if len(queueEntry.Metadata.Artists) > 1 {
			track += " ft. " + "[" + queueEntry.Metadata.Artists[1].Name + "](" + queueEntry.Metadata.Artists[1].URL + ")"
			if len(queueEntry.Metadata.Artists) > 2 {
				for i, artist := range queueEntry.Metadata.Artists[2:] {
					track += ", "
					if (i + 3) == len(queueEntry.Metadata.Artists) {
						track += " and "
					}
					track += "[" + artist.Name + "](" + artist.URL + ")"
				}
			}
		}
	}
	return NewEmbed().
		AddField("Now Playing from "+queueEntry.ServiceName, track).
		AddField("Time", playbackPosition+" / "+fullDuration).
		AddField("Requester", "<@!"+queueEntry.Requester.ID+">").
		SetThumbnail(queueEntry.Metadata.ThumbnailURL).
		SetColor(queueEntry.ServiceColor).MessageEmbed
}
func (queueEntry *QueueEntry) GetQueueAddedEmbed() *discordgo.MessageEmbed {
	track := "[" + queueEntry.Metadata.Title + "](" + queueEntry.Metadata.DisplayURL + ")"
	if len(queueEntry.Metadata.Artists) > 0 {
		track += " by [" + queueEntry.Metadata.Artists[0].Name + "](" + queueEntry.Metadata.Artists[0].URL + ")"
		if len(queueEntry.Metadata.Artists) > 1 {
			track += " ft. " + "[" + queueEntry.Metadata.Artists[1].Name + "](" + queueEntry.Metadata.Artists[1].URL + ")"
			if len(queueEntry.Metadata.Artists) > 2 {
				for i, artist := range queueEntry.Metadata.Artists[2:] {
					track += ", "
					if (i + 3) == len(queueEntry.Metadata.Artists) {
						track += " and "
					}
					track += "[" + artist.Name + "](" + artist.URL + ")"
				}
			}
		}
	}
	return NewEmbed().
		AddField("Added to Queue from "+queueEntry.ServiceName, track).
		AddField("Duration", secondsToHuman(queueEntry.Metadata.Duration)).
		AddField("Requester", "<@!"+queueEntry.Requester.ID+">").
		SetThumbnail(queueEntry.Metadata.ThumbnailURL).
		SetColor(queueEntry.ServiceColor).MessageEmbed
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

	//Setup pointers to guild data for local usage
	//var voiceConnection *discordgo.VoiceConnection = guildData[guildID].VoiceData.VoiceConnection
	//var encodingSession *dca.EncodeSession = guildData[guildID].VoiceData.EncodingSession
	//var streamingSession *dca.StreamingSession = guildData[guildID].VoiceData.StreamingSession

	//Setup the audio encoding options
	if guildData[guildID].VoiceData.EncodingOptions == nil {
		guildData[guildID].VoiceData.EncodingOptions = encodeOptionsPresetHigh
	}

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

func voicePlayWrapper(session *discordgo.Session, guildID, channelID string, queueEntry *QueueEntry, announceQueueAdded bool) {
	if voiceIsStreaming(guildID) {
		guildData[guildID].QueueAdd(queueEntry)
		if announceQueueAdded {
			session.ChannelMessageSendEmbed(channelID, queueEntry.GetQueueAddedEmbed())
		}
		return
	}

	guildData[guildID].AudioNowPlaying = queueEntry
	session.ChannelMessageSendEmbed(channelID, queueEntry.GetNowPlayingEmbed())
	err := voicePlay(guildID, queueEntry.Metadata.StreamURL)
	if guildData[guildID].VoiceData.RepeatLevel == 2 { //Repeat Now Playing
		for guildData[guildID].VoiceData.RepeatLevel == 2 {
			session.ChannelMessageSendEmbed(channelID, queueEntry.GetNowPlayingEmbed())
			err = voicePlay(guildID, queueEntry.Metadata.StreamURL)
			if err != nil && guildData[guildID].VoiceData.IsPlaybackRunning == false {
				guildData[guildID].AudioNowPlaying = nil //Clear now playing slot
				session.ChannelMessageSendEmbed(channelID, NewErrorEmbed("Voice Error", "There was an error playing the specified audio."))
				return
			}
		}
	}
	if guildData[guildID].VoiceData.RepeatLevel == 1 { //Repeat Playlist
		guildData[guildID].QueueAdd(guildData[guildID].AudioNowPlaying) //Shift the now playing entry to the end of the guild queue
	}
	guildData[guildID].AudioNowPlaying = nil //Clear now playing slot
	if err != nil && guildData[guildID].VoiceData.IsPlaybackRunning == false {
		session.ChannelMessageSendEmbed(channelID, NewErrorEmbed("Voice Error", "There was an error playing the specified audio."))
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
				session.ChannelMessageSendEmbed(channelID, guildData[guildID].AudioNowPlaying.GetNowPlayingEmbed())

				//Play audio
				err := voicePlay(guildID, guildData[guildID].AudioNowPlaying.Metadata.StreamURL)
				if guildData[guildID].VoiceData.RepeatLevel == 2 { //Repeat Now Playing
					for guildData[guildID].VoiceData.RepeatLevel == 2 {
						session.ChannelMessageSendEmbed(channelID, guildData[guildID].AudioNowPlaying.GetNowPlayingEmbed())
						err = voicePlay(guildID, guildData[guildID].AudioNowPlaying.Metadata.StreamURL)
						if err != nil && guildData[guildID].VoiceData.IsPlaybackRunning == false {
							guildData[guildID].AudioNowPlaying = nil //Clear now playing slot
							session.ChannelMessageSendEmbed(channelID, NewErrorEmbed("Voice Error", "There was an error playing the specified audio."))
							return
						}
					}
				}
				if guildData[guildID].VoiceData.RepeatLevel == 1 { //Repeat Playlist
					guildData[guildID].QueueAdd(guildData[guildID].AudioNowPlaying) //Shift the now playing entry to the end of the guild queue
				}
				guildData[guildID].AudioNowPlaying = nil //Clear now playing slot
				if err != nil && guildData[guildID].VoiceData.IsPlaybackRunning == false {
					session.ChannelMessageSendEmbed(channelID, NewErrorEmbed("Voice Error", "There was an error playing the specified audio."))
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
				session.ChannelMessageSendEmbed(channelID, NewGenericEmbed("Voice", "Finished playing the queue."))
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

func stopPlaybackPreparing(guildID string) {
	guildData[guildID].VoiceData.IsPlaybackPreparing = false
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

func isSpotifyPlaylistURLURI(url string) bool {
	regexHasSpotify, _ := regexp.MatchString("^(https:\\/\\/open.spotify.com\\/user\\/|spotify:user:)(\\w\\S+)(\\/playlist\\/|:playlist:)(\\w\\S+)(.*)$", url)
	if regexHasSpotify {
		return true
	}
	return false
}
