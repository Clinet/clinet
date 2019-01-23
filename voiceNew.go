package main

import (
	"errors"
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

type RepeatLevel int

const (
	RepeatNone RepeatLevel = iota
	RepeatPlaylist
	RepeatNowPlaying
)

//Voice contains data about the current voice session
type Voice struct {
	//Voice connections and audio sessions
	VoiceConnection  *discordgo.VoiceConnection `json:"voiceConnection"`  //The current Discord voice connection
	EncodingSession  *dca.EncodeSession         `json:"encodingSession"`  //The encoding session for encoding the audio stream to Opus
	StreamingSession *dca.StreamingSession      `json:"streamingSession"` //The streaming session for sending the Opus audio to Discord

	//Voice configurations
	EncodingOptions *dca.EncodeOptions `json:"encodingOptions"` //The settings that will be used for encoding the audio stream to Opus
	RepeatLevel     RepeatLevel        `json:"repeatLevel"`     //0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing
	Shuffle         bool               `json:"shuffle"`         //Whether to continue with a shuffled queue or not
	Muted           bool               `json:"muted"`           //Whether or not audio should be sent to Discord
	Deafened        bool               `json:"deafened"`        //Whether or not audio should be received from Discord

	//Assistant configurations
	AssistantEnabled         bool `json:"assistantEnabled"`         //Whether or not the Google Assistant should be enabled
	AssistantNLP             bool `json:"assistantNLP"`             //Whehter or not to use the Google Assistant transcripts for Clinet's natural language commands
	AssistantSendTranscripts bool `json:"assistantSendTranscripts"` //Whether or not the Google Assistant transcripts should be sent to the channel specified in voice.TextChannelID
	AssistantSendAudio       bool `json:"assistantSendAudio"`       //Whether or not the Google Assistant responses should be played in the current voice channel

	//Contains data about the current queue
	Entries          []*QueueEntry   `json:"queueEntries"`               //Holds a list of queue entries
	ShuffledPointers []int           `json:"shuffledQueueEntryPointers"` //Holds a list of numeric pointers to queue entries for shuffling around freely
	NowPlaying       VoiceNowPlaying `json:"nowPlaying"`                 //Holds the queue entry currently in the now playing slot

	//Miscellaneous
	TextChannelID string `json:"textChannelID"` //The channel that was last used to interact with the voice session
}

func (voice *Voice) Connect(guildID, vChannelID string) error {
	//If a voice connection is already established...
	if voice.IsConnected() {
		//...Change the voice channel
		err := voice.VoiceConnection.ChangeChannel(vchannelID, voice.Muted, voice.Deafened)
		if err != nil {
			//There was an error changing the voice channel
			return errVoiceJoinChangeChannel
		}

		//Changing the voice channel worked out fine
		return nil
	}

	//Join the voice channel
	voiceConnection, err := botData.DiscordSession.ChannelVoiceJoin(guildID, channelID, voice.Muted, voice.Deafened)
	if err != nil {
		//There was an error joining the voice channel
		return errVoiceJoinChannel
	}

	//Store the new voice connection in memory
	voice.VoiceConnection = voiceConnection

	//Start the Google Assistant
	voice.AssistantStart()

	//Joining the voice channel worked out fine
	return nil
}
func (voice *Voice) Disconnect() error {
	//If a voice connection is already established...
	if voice.IsConnected() {
		//...

		//Stop the Google Assistant
		voice.AssistantStop()

		//Leave the voice channel
		err := voice.VoiceConnection.Disconnect()
		if err != nil {
			//There was an error leaving the voice channel
			return errVoiceLeaveChannel
		}

		//Clear the old voice connection in memory
		voice.VoiceConnection = nil

		//Leaving the voice channel worked out fine
		return nil
	}

	//We're not in a voice channel right now
	return errVoiceLeaveNotConnected
}
func (voice *Voice) Play(mediaURL string) error {
	//Placeholder for now
}
func (voice *Voice) Stop() error {
	//Make sure we're streaming first
	if !voice.IsStreaming() {
		return errVoiceNotStreaming
	}

	//Stop the encoding session
	err := voice.EncodingSession.Stop()
	if err != nil {
		return err
	}

	//Clean up the encoding session
	voice.EncodingSession.Cleanup()

	return nil
}
func (voice *Voice) Skip() error {
	//If we're streaming already...
	if voice.IsStreaming() {
		//...Stop the current now playing
		err := voice.Stop()
		if err != nil {
			return err
		}
	}

	switch voice.RepeatLevel {
	case RepeatNone:
	case RepeatPlaylist:
	case RepeatNowPlaying:
	}
}
func (voice *Voice) IsConnected() bool {
	//Return true if a voice connection exists, otherwise return false
	return voice.VoiceConnection != nil
}
func (voice *Voice) IsStreaming() bool {
	//Return false if a voice connection does not exist
	if !voice.IsConnected() {
		return false
	}

	//Return true if a streaming session exists
	if voice.StreamingSession != nil {
		return true
	}

	//Return true if an encoding session exists
	if voice.EncodingSession != nil {
		return true
	}

	//Otherwise return false
	return false
}
func (voice *Voice) SetTextChannel(tChannelID string) {
	//Set voice message output to current text channel
	voice.TextChannelID = tChannelID
}
func (voice *Voice) AssistantStart() {
	//Placeholder for now
}
func (voice *Voice) AssistantStop() {
	//Placeholder for now
}

func VoiceInit(guildID string) {
	if voiceData[guildID] != nil {
		return
	}

	voiceData[guildID] = &Voice{
		EncodingOptions: botData.BotOptions.AudioEncoding,
	}
}
