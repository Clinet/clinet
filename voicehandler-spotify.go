package main

import (
	"regexp"
	"strings"
)

// Spotify exports the methods required to access the Spotify service
type Spotify struct {
}

// GetName returns "Spotify"
func (*Spotify) GetName() string {
	return "Spotify"
}

// GetColor returns 0x1DB954
func (*Spotify) GetColor() int {
	return 0x1DB954
}

// TestURL tests if the given URL is a Spotify track URL
func (*Spotify) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("^(https:\\/\\/open.spotify.com\\/track\\/|spotify:track:)([a-zA-Z0-9]+)(.*)$", url)
	return test, err
}

// GetMetadata returns the metadata for a given Spotify track URL
func (*Spotify) GetMetadata(url string) (*Metadata, error) {
	if strings.HasPrefix(url, "spotify:track:") {
		url = "https://open.spotify.com/track/" + strings.TrimPrefix(url, "spotify:track:")
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
			Name: artist.Title,
			URL:  "https://open.spotify.com/artist/" + artist.ArtistID,
		}
		metadata.Artists = append(metadata.Artists, *trackArtist)
	}

	return metadata, nil
}
