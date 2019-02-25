package main

import (
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/bwmarrin/discordgo"
)

var (
	fnvHash hash.Hash32 = fnv.New32a()

	errVoiceJoinAlreadyInChannel = errors.New("voice: error joining channel, already in selected voice channel")
	errVoiceJoinBusy             = errors.New("voice: error joining channel, busy in another channel")
	errVoiceJoinChannel          = errors.New("voice: error joining channel")
	errVoiceJoinChangeChannel    = errors.New("voice: error changing channel")
	errVoiceLeaveChannel         = errors.New("voice: error leaving channel")
	errVoiceLeaveNotConnected    = errors.New("voice: error leaving channel, not connected")
	errVoiceNotStreaming         = errors.New("voice: not streaming")
	errVoicePausedAlready        = errors.New("voice: already paused")
	errVoicePlayAlreadyStreaming = errors.New("voice: error playing audio, already streaming")
	errVoicePlayInvalidURL       = errors.New("voice: error playing audio, invalid URL")
	errVoicePlayMuted            = errors.New("voice: error playing audio, muted")
	errVoicePlayNotConnected     = errors.New("voice: error playing audio, not connected")
	errVoicePlayingAlready       = errors.New("voice: already playing")
	errVoiceSkippedManually      = errors.New("voice: skipped audio manually")
	errVoiceStoppedManually      = errors.New("voice: stopped audio manually")
)

func getErrorMessage(err error) (errHash, errMsg string) {
	errHash = hashError(err)

	switch err {
	case errVoiceJoinAlreadyInChannel:
		errMsg = "Already connected to the specified voice channel."
	case errVoiceJoinBusy:
		errMsg = "Busy streaming in another voice channel."
	case errVoiceJoinChannel:
		errMsg = "Error joining the voice channel."
	case errVoiceLeaveChannel:
		errMsg = "Error leaving the voice channel."
	case errVoiceLeaveNotConnected:
		errMsg = "Not connected to a voice channel."
	}

	return
}

func getErrorEmbed(title string, err error) *discordgo.MessageEmbed {
	errHash, errMsg := getErrorMessage(err)
	return NewEmbed().
		AddField(title+" Error", errMsg).
		SetFooter("Error Code: " + errHash).
		SetColor(0xFF0000).MessageEmbed
}

func hashError(err error) string {
	fnvHash.Write([]byte(fmt.Sprintf("%v", err)))
	defer fnvHash.Reset()

	return fmt.Sprintf("%x", fnvHash.Sum(nil))
}
