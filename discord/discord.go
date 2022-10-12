package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/JoshuaDoes/logger"
)

var Log *logger.Logger
var Discord *ClientDiscord

//ClientDiscord holds a Discord session
type ClientDiscord struct {
	*discordgo.Session

	User *discordgo.User
}

func StartDiscord(cfg *CfgDiscord) {
	DiscordCfg = cfg
	Log.Trace("--- StartDiscord() ---")

	Log.Debug("Creating Discord struct...")
	discord, err := discordgo.New("Bot " + DiscordCfg.Token)
	if err != nil {
		Log.Fatal("Unable to connect to Discord!")
	}

	//Only enable informational Discord logging if we're tracing
	if Log.Verbosity == 2 {
		Log.Debug("Setting Discord log level to informational...")
		discord.LogLevel = discordgo.LogInformational
	}

	Log.Info("Registering Discord event handlers...")
	discord.AddHandler(discordReady)
	discord.AddHandler(discordMessageCreate)
	discord.AddHandler(discordInteractionCreate)

	Log.Info("Connecting to Discord...")
	err = discord.Open()
	if err != nil {
		Log.Fatal("Unable to connect to Discord!", err)
		return
	}

	Log.Info("Connected to Discord!")
	Discord = &ClientDiscord{discord, nil}

	Log.Info("Recycling old application commands...")
	if oldAppCmds, err := Discord.ApplicationCommands(Discord.State.User.ID, ""); err == nil {
		for _, cmd := range oldAppCmds {
			Log.Trace("Deleting application command for ", cmd.Name)
			if err := Discord.ApplicationCommandDelete(Discord.State.User.ID, "", cmd.ID); err != nil {
				Log.Error(err)
			}
		}
	}

	Log.Info("Registering application commands...")
	Log.Warn("TODO: Batch overwrite commands, then get a list of commands from Discord that aren't in memory and delete them")
	for _, cmd := range CmdsToAppCommands() {
		Log.Trace("Registering cmd: ", cmd)
		_, err := Discord.ApplicationCommandCreate(Discord.State.User.ID, "", cmd)
		if err != nil {
			Log.Fatal(fmt.Sprintf("Unable to register cmd '%s': %v", cmd.Name, err))
		}
	}
	Log.Info("Application commands ready for use!")
}
