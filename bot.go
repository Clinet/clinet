package main

import (
	//std necessities
	"os"
	"os/signal"
	"syscall"
)

//Global error value because functions are mean
var err error

var (
	cfg *Config //Stores the configuration for the bot
)

func doBot() {
	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)
	log.Trace("--- doBot() ---")

	log.Info("Loading configuration...")
	cfg, err = loadConfig(config, ConfigTypeJSON)
	if err != nil {
		log.Error("Error loading configuration: ", err)
	}

	if writeConfigTemplate {
		log.Info("Updating configuration template...")
		var templateCfg *Config = &Config{}
		templateCfg.SaveTo("config.template.json", ConfigTypeJSON)
	}

	//Load modules
	log.Info("Loading modules...")
	loadModules()

	//Start Discord
	log.Info("Starting Discord...")
	startDiscord()
	defer closeDiscord()

	log.Debug("Waiting for SIGINT syscall signal")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	log.Info("Good-bye!")
}