package main

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/JoshuaDoes/goprobe"
)

// SoundCloud exports the methods required to access the SoundCloud service
type SoundCloud struct {
}

// GetName returns "SoundCloud"
func (*SoundCloud) GetName() string {
	return "SoundCloud"
}

// GetColor returns 0xFF7700
func (*SoundCloud) GetColor() int {
	return 0xFF7700
}

// TestURL tests if the given URL is a SoundCloud track URL
func (*SoundCloud) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("^(https?:\\/\\/)?(www.)?(m\\.)?soundcloud\\.com\\/[\\w\\-\\.]+(\\/)+[\\w\\-\\.]+/?$", url)
	if test {
		return true, err
	}
	return false, err
}

// GetMetadata returns the metadata for a given SoundCloud track URL
func (*SoundCloud) GetMetadata(url string) (*Metadata, error) {
	trackInfo, err := botData.BotClients.SoundCloud.GetTrackInfo(url)
	if err != nil {
		return nil, err
	}

	streamURL := trackInfo.DownloadURL

	probe, err := goprobe.ProbeMedia(streamURL)
	if err != nil {
		return nil, err
	}

	if len(probe.Streams) == 0 {
		return nil, errors.New("goprobe: no media streams found")
	}

	audioStream := &goprobe.Stream{}
	for _, stream := range probe.Streams {
		if stream.CodecType == "audio" {
			audioStream = stream
			break
		}
	}
	if audioStream == nil {
		return nil, errors.New("goprobe: no audio streams found")
	}

	duration, err := strconv.ParseFloat(audioStream.Duration, 64)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		Title:        trackInfo.Title,
		DisplayURL:   url,
		StreamURL:    trackInfo.DownloadURL,
		Duration:     duration,
		ArtworkURL:   trackInfo.ArtURL,
		ThumbnailURL: trackInfo.ArtURL,
	}

	trackArtist := &MetadataArtist{
		Name: trackInfo.Artist,
	}
	metadata.Artists = append(metadata.Artists, *trackArtist)

	return metadata, nil
}
