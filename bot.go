package main

import (
	//std necessities
	"os"
	"os/signal"
	"syscall"

	//Logging
	"github.com/JoshuaDoes/logger"

	//Clinet bot framework
	"github.com/Clinet/clinet_bot"
	"github.com/Clinet/clinet_config"
	"github.com/Clinet/clinet_features"

	//Clinet's features
	"github.com/Clinet/clinet_features_dumpctx"
	"github.com/Clinet/clinet_features_essentials"
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
var log    *logger.Logger

func doBot() {
	log = logger.NewLogger("bot", verbosity)
	log.Trace("--- doBot() ---")

	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)

	log.Info("Loading features...")
	cfg, err := config.LoadConfig(featuresFile, config.ConfigTypeJSON)
	if err != nil {
		log.Error("Error loading features: ", err)
		return
	}

	if cfg.Features == nil || len(cfg.Features) == 0 {
		log.Error("No features found in configuration!")
		return
	}

	log.Debug("Syncing format of features file...")
	cfg.SaveTo(featuresFile, config.ConfigTypeJSON)

	//Initialize Clinet using the configuration provided
	log.Info("Initializing instance of Clinet...")
	clinet = bot.NewBot(cfg)

	//Clinet is effectively online at this stage, so defer shutdown in case of errors below
	defer clinet.Shutdown()

	//Register the conversation services to handle queries
	log.Debug("Registering conversation services...")
	log.Trace("- duckduckgo")
	logFatalError(clinet.RegisterFeature(duckduckgo.Feature))
	log.Trace("- wolframalpha")
	logFatalError(clinet.RegisterFeature(wolframalpha.Feature))

	//Register the features to handle commands
	log.Debug("Registering features...")
	log.Trace("- dumpctx")
	logFatalError(clinet.RegisterFeature(dumpctx.Feature))
	log.Trace("- hellodolly")
	logFatalError(clinet.RegisterFeature(hellodolly.Feature))
	log.Trace("- moderation")
	logFatalError(clinet.RegisterFeature(moderation.Feature))
	log.Trace("- voice")
	logFatalError(clinet.RegisterFeature(voice.Feature))
	log.Trace("- essentials")
	logFatalError(clinet.RegisterFeature(essentials.Feature)) //ALWAYS REGISTER ESSENTIALS LAST BEFORE CHAT SERVICES!

	log.Debug("Registering chat services...")
	log.Trace("- discord")
	logFatalError(clinet.RegisterFeature(discord.Feature))
	log.Trace("- guilded")
	logFatalError(clinet.RegisterFeature(guilded.Feature))

	if writeFeaturesTemplate {
		log.Debug("Updating features template...")
		templateCfg := config.NewConfig()
		templateCfg.Features = features.FM.Features
		templateCfg.SaveTo("features.template.json", config.ConfigTypeJSON)
	}

	log.Info("Clinet is now online!")

	log.Debug("Waiting for SIGINT syscall signal...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	log.Info("Good-bye!")
}

func logFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}