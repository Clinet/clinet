package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (voice *Voice) QueueAdd(entry *QueueEntry) {
	//Add the new queue entry
	voice.Entries = append(voice.Entries, entry)
}
func (voice *Voice) QueueRemove(entry int) {
	if voice.Shuffle {
		//Remove the underlying queue entry
		voice.queueRemoveNormal(voice.ShuffledPointers[entry])
		//Remove the pointer from the shuffled queue entries
		voice.queueRemovePointer(entry)
	} else {
		//Remove the queue entry
		voice.queueRemoveNormal(entry)
	}
}
func (voice *Voice) queueRemoveNormal(entry int) {
	voice.Entries = append(voice.Entries[:entry], voice.Entries[entry+1:]...)
}
func (voice *Voice) queueRemovePointer(entry int) {
	voice.ShuffledPointers = append(voice.ShuffledPointers[:entry], voice.ShuffledPointers[entry+1:]...)
}
func (voice *Voice) QueueRemoveRange(start, end int) {
	if len(voice.Entries) == 0 {
		return
	}

	if start < 0 {
		start = 0
	}
	if end > len(voice.Entries) {
		end = len(voice.Entries)
	}

	for entry := end; entry < start; entry-- {
		voice.QueueRemove(entry)
	}
}
func (voice *Voice) QueueClear() {
	voice.Entries = nil
	voice.ShuffledPointers = nil
}
func (voice *Voice) QueueGet(entry int) *QueueEntry {
	if len(voice.Entries) < entry {
		return nil
	}

	if voice.Shuffle {
		return voice.Entries[voice.ShuffledPointers[entry]]
	}
	return voice.Entries[entry]
}
func (voice *Voice) QueueGetNext() *QueueEntry {
	if len(voice.Entries) == 0 {
		return nil
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

func (voice *Voice) GetNowPlayingEmbed(entry *QueueEntry) *discordgo.MessageEmbed {
	return voice.getQueueEmbed(entry, 1)
}
func (voice *Voice) GetNowPlayingDurationEmbed(entry *QueueEntry) *discordgo.MessageEmbed {
	return voice.getQueueEmbed(entry, 2)
}
func (voice *Voice) GetAddedEmbed(entry *QueueEntry) *discordgo.MessageEmbed {
	return voice.getQueueEmbed(entry, 3)
}
func (voice *Voice) getQueueEmbed(entry *QueueEntry, embedType int) *discordgo.MessageEmbed {
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
		embed.AddField("Time", fmt.Sprintf("%s / %s", secondsToHuman(voice.NowPlaying.Position.Seconds()), duration))
	case 3:
		embed.AddField("Added to Queue from "+entry.ServiceName, track)
		embed.AddField("Duration", duration)
	}

	embed.SetColor(entry.ServiceColor)
	embed.SetThumbnail(entry.Metadata.ArtworkURL)

	return embed.MessageEmbed
}

//VoiceNowPlaying contains data about the now playing queue entry
type VoiceNowPlaying struct {
	Entry    *QueueEntry   //The underlying queue entry
	Position time.Duration //The current position in the audio stream
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
