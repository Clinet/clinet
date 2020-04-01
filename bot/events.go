package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

func discordMessageCreate(session *discord.Session, event *discord.MessageCreate) {
	Log.Trace("--- discordMessageCreate(", session, ", ", event, ") ---")
}