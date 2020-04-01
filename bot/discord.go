package bot

import (
	//Discord-related essentials
	discord "github.com/bwmarrin/discordgo" //Used to communicate with Discord

	//std necessities
	"os"
	"os/signal"
	"syscall"
)

var Discord *discord.Session

func startDiscord() {
	Log.Trace("--- startDiscord() ---")

	Log.Debug("Creating Discord struct")
	Discord, err = discord.New("Bot " + cfg.Discord.Token)
	if err != nil {
		Log.Fatal("Unable to connect to Discord!")
	}

	Log.Debug("Deferring closing of Discord")
	defer Discord.Close()

	//Only enable informational Discord logging if we're tracing
	if Log.Verbosity == 2 {
		Log.Debug("Setting Discord log level to informational")
		Discord.LogLevel = discord.LogInformational
	}

	Log.Info("Registering Discord event handlers")
	Discord.AddHandler(discordMessageCreate)

	Log.Info("Connecting to Discord")
	err = Discord.Open()
	if err != nil {
		Log.Fatal("Unable to connect to Discord!")
	}
	Log.Info("Connected to Discord!")

	Log.Debug("Waiting for SIGINT syscall signal")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	Log.Info("Good-bye!")
}