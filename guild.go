package main

import (
	"sync"

	"github.com/JoshuaDoes/go-wolfram"
)

// GuildData holds data specific to a guild
type GuildData struct {
	sync.Mutex //This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

	Queries              map[string]*Query                `json:"queries,omitempty"`
	YouTubeResults       map[string]*YouTubeResultNav     `json:"youtubeResults,omitempty"`
	SpotifyResults       map[string]*SpotifyResultNav     `json:"spotifyResults,omitempty"`
	WolframConversations map[string]*wolfram.Conversation `json:"wolframConversations,omitempty"`
}
