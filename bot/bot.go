package bot

import (
	//Logging
	"github.com/Clinet/clinet/utils/logger" //Advanced logging

	//Discord-related essentials
	//"github.com/bwmarrin/discordgo" //Used to communicate with Discord

	//std necessities
	"os"
)

var Log *logger.Logger

func Bot(log *logger.Logger) {
	defer os.Exit(0)

	Log = log
	Log.Trace("--- Bot() ---")

	Log.Warn("Bot mode is not yet implemented!")
}