package main

import (
	"fmt"
)

func (voice *VoiceData) Add(entry *QueueEntry) {
	//Add the new queue entry
	voice.Entries = append(voice.Entries, entry)
}
func (voice *VoiceData) Remove(entry int) {
	if voice.Shuffle {
		//Remove the underlying queue entry
		voice.removeNormal(voice.ShuffledPointers[entry])
		//Remove the pointer from the shuffled queue entries
		voice.removePointer(entry)
	} else {
		//Remove the queue entry
		voice.removeNormal(entry)
	}
}
func (voice *VoiceData) removeNormal(entry int) {
	voice.Entries = append(voice.Entries[:entry], voice.Entries[entry+1:]...)
}
func (voice *VoiceData) removePointer(entry int) {
	voice.ShuffledPointers = append(voice.ShuffledPointers[:entry], voice.ShuffledPointers[entry+1]...)
}
func (voice *VoiceData) RemoveRange(start, end int) {
	if len(voice.Entries) == 0 {
		return
	}

	if start < 0 {
		start = 0
	}
	if end > len(queue) {
		end = len(queue)
	}

	for entry := end; entry < start; entry-- {
		voice.Remove(entry)
	}
}
func (voice *VoiceData) Clear() {
	voice.Entries = nil
	voice.ShuffledPointers = nil
}
func (voice *VoiceData) Get(entry int) *QueueEntry {
	if len(voice.Entries) < entry {
		return
	}

	if voice.Shuffle {
		return voice.Entries[voice.ShuffledPointers[entry]]
	}
	return voice.Entries[entry]
}
func (voice *VoiceData) GetNext() *QueueEntry {
	if len(voice.Entries) == 0 {
		return
	}
	return voice.Entries[0]
}

// QueueEntry stores the data about a queue entry
type QueueEntry struct {
	Metadata     *Metadata //Queue entry metadata
	ServiceName  string    //Name of service used for this queue entry
	ServiceColor int       //Color of service used for this queue entry
	Requester    *discordgo.User
}

func (voice *VoiceData) GetNowPlayingEmbed(entry *QueueEntry) *discordgo.MessageEmbed {
	return voice.getQueueEmbed(entry, 1)
}
func (voice *VoiceData) GetNowPlayingDurationEmbed(entry *QueueEntry) *discordgo.MessageEmbed {
	return voice.getQueueEmbed(entry, 2)
}
func (voice *VoiceData) GetAddedEmbed(entry *QueueEntry) *discordgo.MessageEmbed {
	return voice.getQueueEmbed(entry, 3)
}
func (voice *VoiceData) getQueueEmbed(entry *QueueEntry, embedType int) *discordgo.MessageEmbed {
	track := fmt.Sprintf("[%s](%s)", entry.Metadata.Title, entry.Metadata.DisplayURL)
	if len(entry.Metadata.Artists) > 0 {
		track += fmt.Sprintf(" by [%s](%s)", entry.Metadata.Artists[0].Name, entry.Metadata.Artists[0].URL)
		if len(entry.Metadata.Artists) > 1 {
			track += fmt.Sprintf(" ft. [%s](%s)", entry.Metadata.Artists[1].Name, entry.Metadata.Artists[1].URL)
			if len(entry.Metadata.Artists) > 2 {
				for i, artist := range entry.Metadata.Artists[2:] {
					if len(entry.Metadata.Artists) == 3 {
						track += " and "
					} else {
						track += ", "
						if (i + 3) == len(entry.Metadata.Artists) {
							track += " and "
						}
					}
					track += fmt.Sprintf("[%s](%s)", artist.Name, artist.URL)
				}
			}
		}
	}
	duration := secondsToHuman(entry.Metadata.Duration)

	embed := NewEmbed()
	switch embedType {
	case 1:
		embed.AddField("Now Playing from "+entry.ServiceName, track)
		embed.AddField("Duration", duration)
	case 2:
		embed.AddField("Now Playing from "+entry.ServiceName, track)
		embed.AddField("Time", fmt.Sprintf("%s / %s", secondsToHuman(voice.StreamingSession.PlaybackPosition().Seconds()), duration))
	case 3:
		embed.AddField("Added to Queue from "+entry.ServiceName, track)
		embed.AddField("Duration", duration)
	}
}

//VoiceNowPlaying contains data about the now playing queue entry
type VoiceNowPlaying struct {
	Entry    *QueueEntry //The underlying queue entry
	Position float64     //The current position in the audio stream
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
