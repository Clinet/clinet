package main

import (
	//std necessities
	"os"
	"os/signal"
	"syscall"

	//Clinet bot framework
	"github.com/Clinet/clinet_bot"
	"github.com/Clinet/clinet_config"
	"github.com/Clinet/clinet_features"

	//Clinet's features
	"github.com/Clinet/clinet_features_dumpctx"
	"github.com/Clinet/clinet_features_hellodolly"
	"github.com/Clinet/clinet_features_moderation"
	"github.com/Clinet/clinet_features_voice"

	//Clinet's services
	"github.com/Clinet/clinet_convos_duckduckgo"
	"github.com/Clinet/clinet_convos_wolframalpha"
	"github.com/Clinet/clinet_services_discord"
	"github.com/Clinet/clinet_services_guilded"
)

var clinet *bot.Bot

func doBot() {
	log.Trace("--- doBot() ---")

	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)

	log.Info("Loading configuration...")
	cfg, err := config.LoadConfig(configFile, config.ConfigTypeJSON)
	if err != nil {
		log.Error("Error loading configuration: ", err)
		return
	}

	log.Info("Syncing configuration...")
	cfg.SaveTo(configFile, config.ConfigTypeJSON)

	if writeConfigTemplate {
		log.Info("Updating configuration template...")
		templateCfg := config.NewConfig()
		templateCfg.Features = append(templateCfg.Features, &features.Feature{Name: "example", Toggle: true})
		templateCfg.SaveTo("config.template.json", config.ConfigTypeJSON)
	}

	clinet = bot.NewBot(cfg)
	defer clinet.Shutdown()

	log.Debug("Registering features...")
	if err := clinet.RegisterFeature(dumpctx.Feature); err != nil {
		log.Fatal(err)
	}
	if err := clinet.RegisterFeature(hellodolly.Feature); err != nil {
		log.Fatal(err)
	}
	if err := clinet.RegisterFeature(moderation.Feature); err != nil {
		log.Fatal(err)
	}
	if err := clinet.RegisterFeature(voice.Feature); err != nil {
		log.Fatal(err)
	}

	if err := clinet.RegisterConvoService("duckduckgo", duckduckgo.DuckDuckGo); err != nil {
		log.Fatal(err)
	}
	if err := clinet.RegisterConvoService("wolframalpha", wolframalpha.WolframAlpha); err != nil {
		log.Fatal(err)
	}

	if err := clinet.RegisterService("discord", discord.Discord); err != nil {
		log.Fatal(err)
	}
	if err := clinet.RegisterService("guilded", guilded.Guilded); err != nil {
		log.Fatal(err)
	}

	log.Debug("Waiting for SIGINT syscall signal...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	log.Info("Good-bye!")
}