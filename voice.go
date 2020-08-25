package main

import (
	"io"
	"net/url"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type RepeatLevel int

const (
	RepeatNone RepeatLevel = iota
	RepeatPlaylist
	RepeatNowPlaying
)

//Voice contains data about the current voice session
type Voice struct {
	sync.Mutex `json:"-"` //This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

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
	Entries          []*QueueEntry    `json:"queueEntries"`               //Holds a list of queue entries
	ShuffledPointers []int            `json:"shuffledQueueEntryPointers"` //Holds a list of numeric pointers to queue entries for shuffling around freely
	NowPlaying       *VoiceNowPlaying `json:"nowPlaying"`                 //Holds the queue entry currently in the now playing slot

	//Miscellaneous
	TextChannelID string     `json:"textChannelID"` //The channel that was last used to interact with the voice session
	done          chan error `json:"-"`             //Used to signal when streaming is done or other actions are performed
}

// Connect connects to a given voice channel
func (voice *Voice) Connect(guildID, vChannelID string) error {
	voice.Lock()
	defer voice.Unlock()

	//If a voice connection is already established...
	if voice.IsConnected() {
		//Change the voice channel
		err := voice.VoiceConnection.ChangeChannel(vChannelID, voice.Muted, voice.Deafened)
		if err != nil {
			//There was an error changing the voice channel
			return errVoiceJoinChangeChannel
		}

		//Changing the voice channel worked out fine
		return nil
	}

	//Join the voice channel
	voiceConnection, err := botData.DiscordSession.ChannelVoiceJoin(guildID, vChannelID, voice.Muted, voice.Deafened)
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

// Disconnect disconnects from the current voice channel
func (voice *Voice) Disconnect() error {
	voice.Lock()
	defer voice.Unlock()

	//If a voice connection is already established...
	if voice.IsConnected() {
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

// Play plays a given queue entry in a connected voice channel
// - queueEntry: The queue entry to play/add to the queue
// - announceQueueAdded: Whether or not to announce a queue added message if something is already playing (used internally for mass playlist additions)
func (voice *Voice) Play(queueEntry *QueueEntry, announceQueueAdded bool) error {
	//Make sure we're conected first
	if !voice.IsConnected() {
		return errVoicePlayNotConnected
	}

	//Make sure we're not streaming already
	if voice.IsStreaming() {
		//If we are streaming, add to the queue instead
		voice.QueueAdd(queueEntry)
		if announceQueueAdded {
			botData.DiscordSession.ChannelMessageSendEmbed(voice.TextChannelID, voice.GetAddedEmbed(queueEntry))
		}
		return nil
	}

	voice.Lock()

	//Make sure we're allowed to speak
	if voice.Muted {
		return errVoicePlayMuted
	}

	//Set the requested entry as now playing
	voice.NowPlaying = &VoiceNowPlaying{Entry: queueEntry}

	//Tell the world we're now playing this entry
	botData.DiscordSession.ChannelMessageSendEmbed(voice.TextChannelID, voice.GetNowPlayingEmbed(queueEntry))

	//Create a channel to signal when the voice stream is finished or stopped
	voice.done = make(chan error)

	voice.Unlock()

	//Start playing this entry
	msg, err := voice.playRaw(voice.NowPlaying.Entry.Metadata.StreamURL)

	if msg != nil {
		if msg == errVoiceStoppedManually {
			return nil
		}
	}

	if err != nil {
		switch err {
		case io.ErrUnexpectedEOF:
			if msg != errVoiceSkippedManually {
				return err
			}
		default:
			return err
		}
	}

	nextQueueEntry := &QueueEntry{}

	switch voice.RepeatLevel {
	case RepeatNone:
		voice.NowPlaying = nil
		if len(voice.Entries) <= 0 {
			voice.Disconnect()
			botData.DiscordSession.ChannelMessageSendEmbed(voice.TextChannelID, NewGenericEmbed("Voice", "Finished playing the queue."))
			return nil
		}
		nextQueueEntry = voice.QueueGet(0)
		voice.QueueRemove(0)
	case RepeatPlaylist:
		voice.QueueAdd(voice.NowPlaying.Entry)
		nextQueueEntry = voice.QueueGet(0)
		voice.QueueRemove(0)
	case RepeatNowPlaying:
		nextQueueEntry = voice.NowPlaying.Entry
	}

	voice.NowPlaying = nil
	return voice.Play(nextQueueEntry, announceQueueAdded)
}

// playRaw plays a given media URL in a connected voice channel
func (voice *Voice) playRaw(mediaURL string) (error, error) {
	/*
		Just in case things change before playRaw is ran, these checks must stay
	*/
	//Make sure we're connected first
	if !voice.IsConnected() {
		return nil, errVoicePlayNotConnected
	}

	//Make sure we're not streaming already
	if voice.IsStreaming() {
		return nil, errVoicePlayAlreadyStreaming
	}

	voice.Lock()

	//Make sure we're allowed to speak
	if voice.Muted {
		return nil, errVoicePlayMuted
	}

	//Ensure that the media URL is valid
	_, err := url.ParseRequestURI(mediaURL)
	if err != nil {
		return nil, errVoicePlayInvalidURL
	}

	//Create the encoding session to encode the audio stream as DCA
	voice.EncodingSession, err = dca.EncodeFile(mediaURL, voice.EncodingOptions)
	if err != nil {
		return nil, err
	}

	//Mark our voice presence as speaking
	voice.Speaking()

	//Create the streaming session to send the encoded DCA audio to Discord
	voice.StreamingSession = dca.NewStream(voice.EncodingSession, voice.VoiceConnection, voice.done)

	voice.Unlock()

	//Start a goroutine to update the current streaming position
	go voice.updatePosition()

	//Wait for the streaming session to finish
	msg := <-voice.done

	voice.Lock()

	//Mark our voice presence as not speaking
	voice.Silent()

	//Figure out why the streaming session stopped
	_, err = voice.StreamingSession.Finished()

	//Clean up the streaming session
	voice.StreamingSession = nil

	//Clean up the encoding session
	voice.EncodingSession.Stop()
	voice.EncodingSession.Cleanup()
	voice.EncodingSession = nil

	voice.Unlock()

	//Return any streaming errors, if any
	return msg, err
}

// updatePosition updates the current position of a playing media
func (voice *Voice) updatePosition() {
	for {
		voice.Lock()

		if voice.StreamingSession == nil || voice.NowPlaying == nil {
			voice.Unlock()
			return
		}
		voice.NowPlaying.Position = voice.StreamingSession.PlaybackPosition()

		voice.Unlock()
	}
}

// Stop stops the playback of a media
func (voice *Voice) Stop() error {
	voice.Lock()
	defer voice.Unlock()

	//Make sure we're streaming first
	if !voice.IsStreaming() {
		return errVoiceNotStreaming
	}

	voice.done <- errVoiceStoppedManually

	//Stop the encoding session
	if err := voice.EncodingSession.Stop(); err != nil {
		return err
	}

	//Clean up the encoding session
	voice.EncodingSession.Cleanup()

	return nil
}

// Skip stops the encoding session of a playing media, allowing the play wrapper to continue to the next media in a queue
func (voice *Voice) Skip() error {
	voice.Lock()
	defer voice.Unlock()

	//Make sure we're streaming first
	if !voice.IsStreaming() {
		return errVoiceNotStreaming
	}

	//Stop the current now playing
	voice.done <- errVoiceSkippedManually

	//Stop the encoding session
	if err := voice.EncodingSession.Stop(); err != nil {
		return err
	}

	//Clean up the encoding session
	voice.EncodingSession.Cleanup()

	return nil
}

// Pause pauses the playback of a media
func (voice *Voice) Pause() (bool, error) {
	voice.Lock()
	defer voice.Unlock()

	//Make sure we're streaming first
	if !voice.IsStreaming() {
		return false, errVoiceNotStreaming
	}

	//Check if we're already paused
	if isPaused := voice.StreamingSession.Paused(); isPaused {
		return true, errVoicePausedAlready
	}

	//Pause the current media playback
	voice.StreamingSession.SetPaused(true)
	return true, nil
}

// Resume resumes the playback of a media
func (voice *Voice) Resume() (bool, error) {
	voice.Lock()
	defer voice.Unlock()

	//Make sure we're streaming first
	if !voice.IsStreaming() {
		return false, errVoiceNotStreaming
	}

	//Check if we're already resumed
	if isPaused := voice.StreamingSession.Paused(); !isPaused {
		return true, errVoicePlayingAlready
	}

	//Resume the current media playback
	voice.StreamingSession.SetPaused(false)
	return true, nil
}

// ToggleShuffle toggles the current shuffle setting and manages the queue accordingly
func (voice *Voice) ToggleShuffle() error {
	return nil
}

// Speaking allows the sending of audio to Discord
func (voice *Voice) Speaking() {
	if voice.IsConnected() {
		voice.VoiceConnection.Speaking(true)
	}
}

// Silent prevents the sending of audio to Discord
func (voice *Voice) Silent() {
	if voice.IsConnected() {
		voice.VoiceConnection.Speaking(false)
	}
}

// IsConnected returns whether or not a voice connection exists
func (voice *Voice) IsConnected() bool {
	if voice.VoiceConnection != nil {
		return true
	}
	return false
}

// IsStreaming returns whether a media is playing
func (voice *Voice) IsStreaming() bool {
	//Return false if a voice connection does not exist
	if !voice.IsConnected() {
		return false
	}

	//Return false if a streaming session does not exist
	if voice.StreamingSession == nil {
		return false
	}

	//Return false if an encoding session does not exist
	if voice.EncodingSession == nil {
		return false
	}

	//Otherwise return true
	return true
}

// SetTextChannel sets the text channel to send messages to
func (voice *Voice) SetTextChannel(tChannelID string) {
	//Set voice message output to current text channel
	voice.TextChannelID = tChannelID
}

// AssistantStart starts the Google Assistant
func (voice *Voice) AssistantStart() {
	//Placeholder for now
}

// AssistantStop shuts down the Google Assistant
func (voice *Voice) AssistantStop() {
	//Placeholder for now
}

// VoiceInit initializes a voice object for the given guild
func VoiceInit(guildID string) {
	if voiceData[guildID] != nil {
		return
	}

	voiceData[guildID] = &Voice{
		EncodingOptions: botData.BotOptions.AudioEncoding,
	}
}
