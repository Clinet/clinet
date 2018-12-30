package main

import (
	"github.com/JoshuaDoes/go-wolfram"
	"sync"
)

// GuildData holds data specific to a guild
type GuildData struct {
	sync.Mutex //This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

	Voice                Voice
	Queries              map[string]*Query
	YouTubeResults       map[string]*YouTubeResultNav
	SpotifyResults       map[string]*SpotifyResultNav
	WolframConversations map[string]*wolfram.Conversation
}
