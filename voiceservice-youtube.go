package main

import (
	"context"
	"errors"
	"regexp"
	"sort"

	isoduration "github.com/channelmeter/iso8601duration"
	//	"github.com/rylio/ytdl"
	youtube "google.golang.org/api/youtube/v3"
)

var (
	bg = context.Background()
)

// VoiceServiceYouTube exports the methods required to access the YouTube service
type VoiceServiceYouTube struct {
}

// GetName returns the service's name
func (*VoiceServiceYouTube) GetName() string {
	return "YouTube"
}

// GetColor returns the service's color
func (*VoiceServiceYouTube) GetColor() int {
	return 0xFF0000
}

// TestURL tests if the given URL is a YouTube video URL
func (*VoiceServiceYouTube) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("(?:https?:\\/\\/)?(?:www\\.)?youtu\\.?be(?:\\.com)?\\/?.*(?:watch|embed)?(?:.*v=|v\\/|\\/)(?:[\\w-_]+)", url)
	return test, err
}

// GetMetadata returns the metadata for a given YouTube video URL
func (*VoiceServiceYouTube) GetMetadata(url string) (*Metadata, error) {
	videoInfo, err := botData.BotClients.YTDL.GetVideoInfo(bg, url)
	if err != nil {
		return nil, err
	}

	formats := videoInfo.Formats
	if len(formats) == 0 {
		return nil, errors.New("unable to get list of formats")
	}

	sort.SliceStable(formats, func(i, j int) bool {
		return formats[i].Itag.Number < formats[j].Itag.Number
	})

	format := formats[0]

	videoURL, err := botData.BotClients.YTDL.GetDownloadURL(bg, videoInfo, format)
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

//YouTubeGetQuery returns YouTube search results
func YouTubeGetQuery(query string) (string, error) {
	call := botData.BotClients.YouTube.Search.List("id").
		Q(query).
		MaxResults(50)

	response, err := call.Do()
	if err != nil {
		return "", errors.New("could not find any results for the specified query")
	}

	for _, item := range response.Items {
		if item.Id.Kind == "youtube#video" {
			url := "https://youtube.com/watch?v=" + item.Id.VideoId
			return url, nil
		}
	}

	return "", errors.New("could not find a video result for the specified query")
}

//VoiceServiceYouTubeResultNav holds YouTube search results, interacted with via commands
type VoiceServiceYouTubeResultNav struct {
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

//Prev goes back to the previous page
func (page *VoiceServiceYouTubeResultNav) Prev() error {
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

//Next goes to the next page
func (page *VoiceServiceYouTubeResultNav) Next() error {
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

//GetResults returns the current search results
func (page *VoiceServiceYouTubeResultNav) GetResults() ([]*youtube.SearchResult, error) {
	if len(page.Results) == 0 {
		return nil, errors.New("No search results found")
	}
	return page.Results, nil
}

//Search starts a search and stores the results
func (page *VoiceServiceYouTubeResultNav) Search(query string) error {
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
