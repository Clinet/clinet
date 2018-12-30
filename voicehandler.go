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

	//Planned for future rewriting of things
	//Search(query string) (results *SearchResults, err error)
}

func initVoiceServices() {
	botData.VoiceServices = make([]VoiceService, 0)

	botData.VoiceServices = append(botData.VoiceServices, &YouTube{})
	botData.VoiceServices = append(botData.VoiceServices, &SoundCloud{})
	botData.VoiceServices = append(botData.VoiceServices, &Spotify{})
	botData.VoiceServices = append(botData.VoiceServices, &Bandcamp{})
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
				ServiceName:  service.Name,
				ServiceColor: service.Color,
			}
			return queueEntry, nil
		}
	}

	return nil, errors.New("error finding service to handle url")
}
