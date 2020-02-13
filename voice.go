package main

import (
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"sync"
	"time"

	"github.com/Clinet/ffgoconv"
	assistant "github.com/JoshuaDoes/google-assistant"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type RepeatLevel int

const (
	RepeatNone RepeatLevel = iota
	RepeatPlaylist
	RepeatNowPlaying
)

var (
	SILENCE = []byte{0xF8, 0xFF, 0xFE}
)

//Voice contains data about the current voice session
type Voice struct {
	sync.Mutex `json:"-"` //This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

	//Voice connections and audio sessions
	VoiceConnection  *discordgo.VoiceConnection `json:"voiceConnection"`  //The current Discord voice connection
	EncodingSession  *dca.EncodeSession         `json:"encodingSession"`  //The encoding session for encoding the audio stream to Opus
	StreamingSession *dca.StreamingSession      `json:"streamingSession"` //The streaming session for sending the Opus audio to Discord

	//Audio processing handled by ffgoconv
	MediaStreamer             *ffgoconv.Streamer   `json:-` //The media streamer for current media playback
	TTSStreamer               *ffgoconv.Streamer   `json:-` //The TTS streamer for playing back text to speech messages
	AssistantPlaybackStreamer *ffgoconv.Streamer   `json:-` //The Google Assistant streamer for playing back Assistant responses
	//AssistantRecordStreamer   *ffgoconv.Streamer   `json:-` //The recording streamer used to record the users in a voice chat
	AssistantVoiceRecorder    *Opus                `json:-` //The recording session used to record users in a voice chat for the Assistant
	Transmuxer                *ffgoconv.Transmuxer `json:-` //The transmuxing session for transmuxing multiple audio streams

	//Voice configurations
	EncodingOptions *dca.EncodeOptions `json:"encodingOptions"` //The settings that will be used for encoding the audio stream to Opus
	RepeatLevel     RepeatLevel        `json:"repeatLevel"`     //0 = No Repeat, 1 = Repeat Playlist, 2 = Repeat Now Playing
	Shuffle         bool               `json:"shuffle"`         //Whether to continue with a shuffled queue or not
	Muted           bool               `json:"muted"`           //Whether or not audio should be sent to Discord
	Deafened        bool               `json:"deafened"`        //Whether or not audio should be received from Discord

	//Google Assistant
	Assistant           *assistant.Assistant `json:-` //The Google Assistant itself
	AssistantIsSpeaking bool                 `json:-` //Whether or not the Google Assistant is speaking

	//Google Assistant configuration
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
		//Stop the Google Assistant
		voice.AssistantStop()

		//Change the voice channel
		err := voice.VoiceConnection.ChangeChannel(vChannelID, voice.Muted, voice.Deafened)
		if err != nil {
			//There was an error changing the voice channel
			return errVoiceJoinChangeChannel
		}
	} else {
		//Join the voice channel
		voiceConnection, err := botData.DiscordSession.ChannelVoiceJoin(guildID, vChannelID, voice.Muted, voice.Deafened)
		if err != nil {
			//There was an error joining the voice channel
			return errVoiceJoinChannel
		}

		//Store the new voice connection in memory
		voice.VoiceConnection = voiceConnection

		//Add a handler for voice speaking updates
		voice.VoiceConnection.AddHandler(discordVoiceSpeakingUpdate)
	}

	//Start the Google Assistant
	if voice.AssistantEnabled {
		err := voice.AssistantStart()
		if err != nil {
			return err
		}
	}

	//Create the transmuxing session to encode the audio stream as MP3 so that dca can understand it
	err := voice.CreateTransmuxer()
	if err != nil {
		return err
	}

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

		//Clean up the transmuxing session
		if voice.MediaStreamer != nil {
			voice.MediaStreamer.Close()
			voice.MediaStreamer = nil
		}
		if voice.TTSStreamer != nil {
			voice.TTSStreamer.Close()
			voice.TTSStreamer = nil
		}
		if voice.AssistantPlaybackStreamer != nil {
			voice.AssistantPlaybackStreamer.Close()
			voice.AssistantPlaybackStreamer = nil
		}
		if voice.Transmuxer != nil {
			voice.Transmuxer.Close()
			voice.Transmuxer = nil
		}

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
	Debug.Println("ASDFPLAY")
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
	Debug.Println("ASDF")
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
	defer voice.Unlock()

	//Make sure we're allowed to speak
	if voice.Muted {
		return nil, errVoicePlayMuted
	}

	//Ensure that the media URL is valid
	_, err := url.ParseRequestURI(mediaURL)
	if err != nil {
		return nil, errVoicePlayInvalidURL
	}
	
	Debug.Println("ADDING MEDIA STREAMER")
	voice.MediaStreamer, err = voice.Transmuxer.AddStreamer(mediaURL, nil, 1.0)
	if err != nil {
		return nil, err
	}

	Debug.Println("CREATING ENCODING SESSION")
	//Create the encoding session to encode the audio stream as DCA
	//voice.EncodingSession, err = dca.EncodeFile(mediaURL, voice.EncodingOptions)
	voice.EncodingSession, err = dca.EncodeMem(voice.Transmuxer, voice.EncodingOptions)
	if err != nil {
		return nil, err
	}

	Debug.Println("SPEAKING")
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
	
	//Clean up the media streamer
	voice.MediaStreamer.Close()
	voice.MediaStreamer = nil

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
	return voice.VoiceConnection != nil
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

// CreateTransmuxer initializes and starts a transmuxing process
func (voice *Voice) CreateTransmuxer() error {
	if voice.Transmuxer == nil {
		var err error
		voice.Transmuxer, err = ffgoconv.NewTransmuxer(nil, "pipe:1", "libmp3lame", "mp3", "320k", 1)
		if err != nil {
			return err
		}

		go voice.Transmuxer.Run()
	}

	return nil
}

// StartListening sends the SILENCE Opus packet to Discord so that we can listen to audio from other users
func (voice *Voice) StartListening() {
	voice.Speaking()
	Debug.Println("Sending SILENCE packet...")
	voice.VoiceConnection.OpusSend <- SILENCE
	//voice.Silent()
	go voice.AssistantListen()
}

// AssistantStart starts the Google Assistant
func (voice *Voice) AssistantStart() error {
	var err error

	if voice.AssistantEnabled {
		Debug.Println("Creating the assistant struct...")
		voice.Assistant = assistant.NewAssistant()
		voice.Assistant.GCPAuth = &assistant.GCPAuthWrapper{PermissionCode: botData.AssistantPermissionCode}

		Debug.Println("Initializing the assistant...")
		//err = voice.Assistant.InitializeRaw(nil, botData.BotOptions.Assistant.AudioBuffer, botData.BotOptions.Assistant.Credentials, nil, "", nil)
		err = voice.Assistant.Initialize(botData.BotOptions.Assistant.Credentials, nil)
		if err != nil {
			return err
		}

		if botData.AssistantPermissionCode == "" {
			Debug.Println("Asking for first OAuth2 sign-in...")
			ownerPrivChannel, err := botData.DiscordSession.UserChannelCreate(botData.BotOwnerID)
			if err != nil {
				Debug.Printf("Error creating private channel with bot owner: %v\n", err)
			} else {
				ownerPrivChannelID := ownerPrivChannel.ID
				botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, "Please open the following URL to authenticate the Google Assistant for the first time: "+voice.Assistant.GCPAuth.AuthURL)
			}
			Info.Println("Please open the following URL to authenticate the Google Assistant for the first time: "+voice.Assistant.GCPAuth.AuthURL)

			Debug.Println("Waiting for permission code...")
			for {
				if code := voice.Assistant.GCPAuth.PermissionCode; code != "" {
					Debug.Println("Storing authentication permission code:", code)
					botData.AssistantPermissionCode = code
					stateSaveAll()
					break
				}
			}
		}

		Debug.Println("Waiting 2 seconds before starting the Assistant...")
		time.Sleep(2 * time.Second)

		Debug.Println("Starting the assistant")
		err = voice.Assistant.Start()
		if err != nil {
			Debug.Println("Trying again...")
			err = voice.Assistant.Start()
			if err != nil {
				return err
			}
		}

		Debug.Println("Starting a listening thread...")
		voice.StartListening()
	}

	return nil
}

func (voice *Voice) AssistantListen() {
	var err error

	silenceCounter := 0

	/*voice.AssistantRecordStreamer, err = ffgoconv.NewStreamer("-", []string{
		"-hide_banner",
		"-stats",
		"-ac", "1",
		"-re", "-i", "-",
		"-acodec", "pcm_s16le",
		"-f", "s16le",
		"-vol", "256",
		"-ar", "16000",
		"-ac", "1",
		"-b:a", "16k",
		"-threads", "1",
		"pipe:1",
	}, 1.0)
	if err != nil {
		return
	}
	Debug.Println(err, voice.AssistantRecordStreamer.Err())*/
	
	//Audio samples from Discord users are stereo 48000Hz
	voice.AssistantVoiceRecorder, err = NewOpus(48000, 2)
	if err != nil {
		Error.Println(err)
		return
	}

	for voice.IsConnected() {
		select {
		case packet := <-voice.VoiceConnection.OpusRecv:
			//Debug.Println("Received a voice packet:", packet.Opus)
			
			if len(packet.Opus) == 3 && packet.Opus[0] == SILENCE[0] && packet.Opus[1] == SILENCE[1] && packet.Opus[2] == SILENCE[2] {
				Debug.Println("Received a silence voice packet")
				//voice.AssistantVoiceRecorder.Write(packet.Opus)
				silenceCounter += 1
				
				if silenceCounter == 10 {
					Debug.Println("Resetting silence counter...")
					silenceCounter = 0
			
					Debug.Println("Playing response...")
					voice.AssistantPlayResponse()
				}
				
				continue
			}

			Debug.Println("Writing new Opus packet to Assistant voice recorder...")
			voice.AssistantVoiceRecorder.Write(packet.Opus)
			
			Debug.Println("Reading back converted audio data from Assistant record streamer...")
					recAudio, err := ioutil.ReadAll(voice.AssistantVoiceRecorder)
					if err != nil {
						Error.Println(err)
						continue
					}

					Debug.Println("Writing converted audio data to Assistant audio in...")
					err = voice.AssistantAudioIn(recAudio)
					if err != nil {
						Error.Println(err)
						continue
					}
		default:
		}
	}

	Debug.Println("Stopping the listening thread...")
}

func (voice *Voice) AssistantAudioIn(audioIn []byte) error {
	loop := int(math.Ceil(float64(len(audioIn)) / float64(8192)))
	for i := 1; i < (loop + 1); i++ {
		Debug.Println("Sending pass", i, "/", loop)

		low := (i - 1) * 8192
		high := i * 8192
		if high > len(audioIn) {
			high = len(audioIn)
		}

		err := voice.Assistant.AudioIn(audioIn[low:high])
		if err != nil {
			return err
		}

		time.Sleep(25 * time.Millisecond)
	}
	return nil
}

func (voice *Voice) AssistantPlayResponse() {
	var err error
	
	Debug.Println("Creating Assistant playback streamer...")
	voice.AssistantPlaybackStreamer, err = ffgoconv.NewStreamer("pipe:", []string{
		"-hide_banner",
		"-stats",
		"-acodec", "pcm_s16le",
		"-f", "s16le",
		"-ar", "16000",
		"-ac", "1",
		"-re", "-i", "-",
		"-acodec", "pcm_f64le",
		"-f", "f64le",
		"-vol", "256",
		"-ar", "48000",
		"-ac", "2",
		"-b:a", "320k",
		"-threads", "1",
		"pipe:1",
	}, 1.0)
	if err != nil {
		Error.Println(err)
		return
	}
	
	wroteAudio := false
	reqTxt := ""
	resTxt := ""
	
	Debug.Println("Starting record response loop...")
	for {
		Debug.Println("Requesting response piece...")
		response, err := voice.Assistant.RequestResponse()
		if err == io.EOF {
			Debug.Println("Received EOF on audio")
			wroteAudio = true
		} else if err != nil {
			Error.Println(err)
			break
		}
		//Debug.Println("Response piece:", response)
		Debug.Println("Getting result...")
		result := response.GetResult()
		//Debug.Println("Result:", result)
		
		if response == nil {
			Debug.Println("Empty response piece")
			continue
		}
		if audioOut := response.GetAudioOut(); audioOut != nil {
			if wroteAudio == false {
				Debug.Println("Writing audio piece to Assistant streamer...")
				voice.AssistantPlaybackStreamer.Write(audioOut.GetAudioData())
			}
		}
		if requestText := result.GetSpokenRequestText(); requestText != "" {
			Debug.Println("Received request text piece:", requestText)
			reqTxt = requestText
		}
		if responseText := result.GetSpokenResponseText(); responseText != "" {
			Debug.Println("Received response text piece:", responseText)
			resTxt = responseText
		}
		if wroteAudio && reqTxt != "" && resTxt != "" {
			Debug.Println("Finished receiving response pieces")
			break
		}
	}

	Debug.Println("Closing input stream...")
	err = voice.Assistant.Conversation.CloseSend()
	if err != nil {
		Error.Println(err)
	}
	
	Debug.Println("Adding Assistant streamer to transmuxing session...")
	_, err = voice.Transmuxer.AddRunningStreamer(voice.AssistantPlaybackStreamer)
	if err != nil {
		Error.Println(err)
	}
	
	Debug.Println("Telling the world what we did!")
	botData.DiscordSession.ChannelMessageSendEmbed(voice.TextChannelID, NewGenericEmbed("Google Assistant", "**Request**\n" + reqTxt + "\n\n**Response**\n" + resTxt))
}

// AssistantStop shuts down the Google Assistant
func (voice *Voice) AssistantStop() {
	if voice.Assistant != nil {
		Debug.Println("Stopping the assistant")
		voice.Assistant.Close()
	}
	voice.AssistantEnabled = false
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

// Event for speaking updates
func discordVoiceSpeakingUpdate(voiceConnection *discordgo.VoiceConnection, voiceSpeaking *discordgo.VoiceSpeakingUpdate) {
	Debug.Printf("Voice speaking update received:\nVoice Speaking: %v", voiceSpeaking)
}
