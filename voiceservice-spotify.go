package main

import (
	"errors"
	"math"
	"regexp"
	"strings"

	"github.com/JoshuaDoes/spotigo"
)

// VoiceServiceSpotify exports the methods required to access the Spotify service
type VoiceServiceSpotify struct {
}

// GetName returns the service's name
func (*VoiceServiceSpotify) GetName() string {
	return "Spotify"
}

// GetColor returns the service's color
func (*VoiceServiceSpotify) GetColor() int {
	return 0x1DB954
}

// TestURL tests if the given URL is a Spotify track URL
func (*VoiceServiceSpotify) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("^(https:\\/\\/open.spotify.com\\/track\\/|spotify:track:)([a-zA-Z0-9]+)(.*)$", url)
	return test, err
}

// GetMetadata returns the metadata for a given Spotify track URL
func (*VoiceServiceSpotify) GetMetadata(url string) (*Metadata, error) {
	if strings.HasPrefix(url, "spotify:track:") {
		newURL := "https://open.spotify.com/track/" + strings.TrimPrefix(url, "spotify:track:")
		url = newURL
	}

	trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(url)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		Title:        trackInfo.Title,
		DisplayURL:   url,
		StreamURL:    trackInfo.StreamURL,
		Duration:     float64(trackInfo.Duration / 1000),
		ArtworkURL:   trackInfo.ArtURL,
		ThumbnailURL: trackInfo.ArtURL,
	}

	for _, artist := range trackInfo.Artists {
		trackArtist := &MetadataArtist{
			Name: artist.Name,
			URL:  "https://open.spotify.com/artist/" + artist.ArtistID,
		}
		metadata.Artists = append(metadata.Artists, *trackArtist)
	}

	return metadata, nil
}

//VoiceServiceSpotifyResultNav returns Spotify search results, interacted with via commands
type VoiceServiceSpotifyResultNav struct {
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

//GetResults returns the current results
func (page *VoiceServiceSpotifyResultNav) GetResults() ([]spotigo.SpotigoSearchHit, error) {
	if page == nil {
		return nil, errors.New("Page not initialized")
	}

	if len(page.Results) == 0 {
		return nil, errors.New("No search results found")
	}
	return page.Results, nil
}

//Search starts a search and stores the search results
func (page *VoiceServiceSpotifyResultNav) Search(query string) error {
	if page == nil {
		return errors.New("Page not initialized")
	}

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

//Playlist loads a Spotify playlist into the results
func (page *VoiceServiceSpotifyResultNav) Playlist(url string) error {
	if page == nil {
		return errors.New("Page not initialized")
	}

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

//Prev goes back to the previous page
func (page *VoiceServiceSpotifyResultNav) Prev() error {
	if page == nil {
		return errors.New("Page not initialized")
	}

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

//Next goes to the next page
func (page *VoiceServiceSpotifyResultNav) Next() error {
	if page == nil {
		return errors.New("Page not initialized")
	}

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

//Jump jumps to a page number
func (page *VoiceServiceSpotifyResultNav) Jump(pageNumber int) error {
	if page == nil {
		return errors.New("Page not initialized")
	}

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
