package bot

import (
	//Logging
	"github.com/Clinet/clinet/utils/logger" //Advanced logging

	//std necessities
	"os"
)

//Global error value because functions are mean
var err error

var (
	Log *logger.Logger //Stores the logger for the session
	cfg *Config
)

func Bot(cfgPath, token string, log *logger.Logger) {
	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)

	Log = log
	Log.Trace("--- Bot(", cfgPath, ", [REDACTED], ", log, ") ---")

	Log.Info("Loading configuration...")
	cfg, err = NewConfig(cfgPath, ConfigTypeJSON)
	if err != nil {
		Log.Error("Error loading configuration: ", err)
	}

	if token != "" {
		Log.Debug("Patching configuration with new token")
		cfg.Discord.Token = token
	}

	startDiscord()

	Log.Info("Good-bye!")
}