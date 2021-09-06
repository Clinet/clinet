package main

import (
	//std necessities
	"os"
	"os/signal"
	"syscall"

	"github.com/Clinet/clinet/discord"
	"github.com/Clinet/clinet/config"
)

//Global error value because functions are mean
var err error

var (
	cfg     *config.Config
	Discord *discord.DiscordSession
)

func doBot() {
	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)
	log.Trace("--- doBot() ---")

	log.Info("Loading configuration...")
	cfg, err = config.LoadConfig(configFile, config.ConfigTypeJSON)
	if err != nil {
		log.Error("Error loading configuration: ", err)
	}

	if writeConfigTemplate {
		log.Info("Updating configuration template...")
		var templateCfg *config.Config = &config.Config{}
		templateCfg.SaveTo("config.template.json", config.ConfigTypeJSON)
	}

	//Load modules
	log.Info("Loading modules...")
	loadModules()

	//Start Discord
	log.Info("Starting Discord...")
	Discord = discord.StartDiscord(cfg.Discord)
	defer Discord.Close()

	log.Debug("Waiting for SIGINT syscall signal")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	log.Info("Good-bye!")
}