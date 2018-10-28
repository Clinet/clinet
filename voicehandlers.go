package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// VoiceService is the interface that wraps the methods required to access and identify a service.
/*
* GetName returns the name of the service. It returns the name of the service to display to human users.
* GetColor returns the color of the service. It returns the color of the service to display to human users.
* TestURL tests if the given URL is valid for the service. It returns a bool dictating whether or not it is valid and any error that caused the test to stop early. If an error is returned, do not trust that the match is valid.
* GetMetadata gets all available metadata for the given URL. It returns all collected metadata and any error that caused the metadata search to stop early. If an error is returned, do not trust that the returned metadata is valid.
 */
type VoiceService interface {
	GetName() string
	GetColor() int
	TestURL(url string) (match bool, err error)
	GetMetadata(url string) (metadata *Metadata, err error)
}

// Metadata stores the metadata of a queue entry
type Metadata struct {
	Artists      []MetadataArtist //List of artists for this queue entry
	Title        string           //Entry title
	DisplayURL   string           //Entry page URL to display to users
	StreamURL    string           //Entry URL for streaming
	Duration     float64          //Entry duration
	ArtworkURL   string           //Entry artwork URL
	ThumbnailURL string           //Entry artwork thumbnail URL
}

// MetadataArtist stores the data about an artist
type MetadataArtist struct {
	Name string //Artist name
	URL  string //Artist page URL
}

// QueueEntry stores the data about a queue entry
type QueueEntry struct {
	Metadata     *Metadata //Queue entry metadata
	ServiceName  string    //Name of service used for this queue entry
	ServiceColor int       //Color of service used for this queue entry
	Requester    *discordgo.User
}

func initVoiceServices() {
	botData.VoiceServices = make([]VoiceService, 0)

	botData.VoiceServices = append(botData.VoiceServices, &YouTube{})
	botData.VoiceServices = append(botData.VoiceServices, &SoundCloud{})
	botData.VoiceServices = append(botData.VoiceServices, &Spotify{})
	botData.VoiceServices = append(botData.VoiceServices, &Direct{})
}

func createQueueEntry(url string) (*QueueEntry, error) {
	for _, service := range botData.VoiceServices {
		test, err := service.TestURL(url)
		if err != nil {
			return nil, err
		}
		if test {
			metadata, err := service.GetMetadata(url)
			if err != nil {
				return nil, err
			}
			queueEntry := &QueueEntry{
				Metadata:     metadata,
				ServiceName:  service.GetName(),
				ServiceColor: service.GetColor(),
			}
			return queueEntry, nil
		}
	}
	return nil, errors.New("error finding service to handle url")
}
