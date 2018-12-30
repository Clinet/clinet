package main

import (
	"regexp"

	isoduration "github.com/channelmeter/iso8601duration"
	"github.com/rylio/ytdl"
	youtube "google.golang.org/api/youtube/v3"
)

// YouTube exports the methods required to access the YouTube service
type YouTube struct {
}

// GetName returns the service's name
func (*YouTube) GetName() string {
	return "YouTube"
}

// GetColor returns the service's color
func (*YouTube) GetColor() int {
	return 0xFF0000
}

// TestURL tests if the given URL is a YouTube video URL
func (*YouTube) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("(?:https?:\\/\\/)?(?:www\\.)?youtu\\.?be(?:\\.com)?\\/?.*(?:watch|embed)?(?:.*v=|v\\/|\\/)(?:[\\w-_]+)", url)
	return test, err
}

// GetMetadata returns the metadata for a given YouTube video URL
func (*YouTube) GetMetadata(url string) (*Metadata, error) {
	videoInfo, err := ytdl.GetVideoInfo(url)
	if err != nil {
		return nil, err
	}

	format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]

	videoURL, err := videoInfo.GetDownloadURL(format)
	if err != nil {
		return nil, err
	}

	ytCall := youtube.NewVideosService(botData.BotClients.YouTube).
		List("snippet,contentDetails").
		Id(videoInfo.ID)

	ytResponse, err := ytCall.Do()
	if err != nil {
		return nil, err
	}

	duration, err := isoduration.FromString(ytResponse.Items[0].ContentDetails.Duration)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		Title:        videoInfo.Title,
		DisplayURL:   url,
		StreamURL:    videoURL.String(),
		Duration:     duration.ToDuration().Seconds(),
		ArtworkURL:   videoInfo.GetThumbnailURL("maxresdefault").String(),
		ThumbnailURL: videoInfo.GetThumbnailURL("default").String(),
	}

	videoAuthor := &MetadataArtist{
		Name: ytResponse.Items[0].Snippet.ChannelTitle,
		URL:  "https://youtube.com/channel/" + ytResponse.Items[0].Snippet.ChannelId,
	}
	metadata.Artists = append(metadata.Artists, *videoAuthor)

	return metadata, nil
}
