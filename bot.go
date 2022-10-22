package main

import (
	//std necessities
	"os"
	"os/signal"
	"syscall"

	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/clinet/config"
	"github.com/Clinet/clinet/convos"
	"github.com/Clinet/clinet/features"
	"github.com/Clinet/clinet/features/dumpctx"
	"github.com/Clinet/clinet/features/hellodolly"
	"github.com/Clinet/clinet/features/moderation"
	"github.com/Clinet/clinet/services/discord"
	"github.com/JoshuaDoes/go-wolfram"
)

//Global error value because functions are mean
var err error

var (
	cfg *config.Config
)

func doBot() {
	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)

	//Assign the logger to each package
	config.Log = log
	discord.Log = log

	log.Trace("--- doBot() ---")

	log.Info("Loading configuration...")
	cfg, err = config.LoadConfig(configFile, config.ConfigTypeJSON)
	if err != nil {
		log.Error("Error loading configuration: ", err)
		return
	}

	log.Info("Syncing configuration...")
	cfg.SaveTo(configFile, config.ConfigTypeJSON)

	if writeConfigTemplate {
		log.Info("Updating configuration template...")
		var templateCfg *config.Config = &config.Config{
			Features: []*features.Feature{&features.Feature{Name: "example", Toggle: true}},
			Discord: &discord.CfgDiscord{},
			WolframAlpha: &wolfram.Client{},
		}
		templateCfg.SaveTo("config.template.json", config.ConfigTypeJSON)
	}

	log.Debug("Setting feature toggles...")
	features.SetFeatures(cfg.Features)

	log.Debug("Registering features...")
	if features.IsEnabled("dumpctx") {
		cmds.Commands = append(cmds.Commands, dumpctx.CmdRoot)
	}
	if features.IsEnabled("hellodolly") {
		cmds.Commands = append(cmds.Commands, hellodolly.CmdRoot)
	}
	if features.IsEnabled("moderation") {
		err = moderation.Init(log)
		if err != nil {
			log.Fatal(err)
		}
		log.Trace(moderation.Storage)
		cmds.Commands = append(cmds.Commands, moderation.CmdRoot)
	}

	log.Debug("Enabling services...")
	if features.IsEnabled("wolframalpha") {
		convos.AuthWolframAlpha(cfg.WolframAlpha)
		log.Trace("- Wolfram|Alpha")
	}

	//Load modules
	log.Info("Loading modules...")
	loadModules()

	//Start Discord
	log.Info("Starting Discord...")
	discord.StartDiscord(cfg.Discord)
	defer discord.Discord.Close()

	log.Debug("Waiting for SIGINT syscall signal...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	log.Info("Good-bye!")
}