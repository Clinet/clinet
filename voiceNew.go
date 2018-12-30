package main

import (
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

//Voice contains everything related to voice activities for a particular guild
type Voice struct {
	Data   VoiceData
	Config VoiceConfig
	Queue  VoiceQueue
}

func VoiceNew() *Voice {
	return &Voice{
		Config: {
			EncodingOptions: audioEncodingPreset,
		},
	}
}

//VoiceData contains data about the current voice session
type VoiceData struct {
	//Voice sockets and stuff, don't save this to disk
	VoiceConnection  *discordgo.VoiceConnection `json:"-"`
	EncodingSession  *dca.EncodeSession         `json:"-"`
	StreamingSession *dca.StreamingSession      `json:"-"`

	//Voice configurations
	EncodingOptions *dca.EncodeOptions
	RepeatLevel     int  //0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing
	Shuffle         bool //Whether to continue with a shuffled queue or not

	//Contains data about the current queue
	Entries          []*QueueEntry   //Holds a list of queue entries
	ShuffledPointers []int           //Holds a list of numeric pointers to queue entries for shuffling around freely
	NowPlaying       VoiceNowPlaying //Holds the queue entry currently in the now playing slot

	//Miscellaneous
	ChannelIDJoinedFrom string
}

//DCA encoding preset
var audioEncodingPreset = &dca.EncodeOptions{
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

func voiceJoin(session *discordgo.Session, guild, voiceChannel, textChannel string) error {
	_, guildFound := guildData[guild]
	if guildFound {
		if guildData[guild].Voice.Data.VoiceConnection != nil {
			if guildData[guild].Voice.Data.VoiceConnection.ChannelID == voiceChannel {
				return errVoiceJoinAlreadyInChannel
			} else {
				if guildData[guild].Voice.Data.StreamingSession != nil {
					return errVoiceJoinBusy
				} else {
					err := voiceLeave(guild, voiceChannel)
					if err != nil {
						return errVoiceLeaveChannel
					}
				}
			}
		}
	} else {
		guildData[guild] = &GuildData{}
		guildData[guild].Voice = &Voice{}
	}

	voiceConnection, err := session.ChannelVoiceJoin(guild, voiceChannel, false, true)
	if err != nil {
		return errVoiceJoinChannel
	}

	guildData[guild].Voice.Data.VoiceConnection = voiceConnection
	guildData[guild].Voice.Data.ChannelIDJoinedFrom = textChannel

	return nil
}

func voiceLeave(guild string) error {
	_, guildFound := guildData[guild]
	if guildFound {
		if guildData[guild].Voice.Data.VoiceConnection != nil {
			guildData[guild].Voice.Data.VoiceConnection.Disconnect()
			guildData[guild].Voice.Data.VoiceConnection = nil
			return nil
		} else {
			return errVoiceLeaveNotConnected
		}
	} else {
		return errVoiceLeaveNotConnected
	}
}

func voiceIsStreaming(guild string) bool {
	if guildData[guild] == nil {
		return false
	}
	if guildData[guild].Voice.Data.VoiceConnection == nil {
		return false
	}
	if guildData[guild].Voice.Data.StreamingSession == nil {
		return false
	}
	return true
}
