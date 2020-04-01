package bot

import (
	//Logging
	"github.com/Clinet/clinet/utils/logger" //Advanced logging

	//Discord-related essentials
	//"github.com/bwmarrin/discordgo" //Used to communicate with Discord

	//std necessities
	"os"
)

//Global error value because functions are mean
var err error

var (
	Log *logger.Logger //Stores the logger for the session
	cfg *Config
)

func Bot(cfgPath string, log *logger.Logger) {
	//For some reason we don't automatically exit as planned when we return to main()
	defer os.Exit(0)

	Log = log
	Log.Trace("--- Bot() ---")

	Log.Info("Loading configuration...")
	cfg, err = NewConfig(cfgPath, ConfigTypeJSON)
	if err != nil {
		Log.Error("Error loading configuration: ", err)
	}
	Log.Debug("cfg: ", cfg)
}