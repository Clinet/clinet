package main

import (
	//The core essentials for Clinet
	"github.com/Clinet/clinet/bot" //Holds the essential bot things

	//Logging
	"github.com/Clinet/clinet/utils/logger" //Advanced logging

	//Additional things
	flag "github.com/spf13/pflag" //Unix-like command-line flags
)

var (
	//Various command-line flags
	cfgPath     string //Stores the path to the configuration file
	token       string //Token override for live streams when you can't show the token
	verbosity   int    //0 = default (info, warning, error), 1 = 0 + debug, 2 = 1 + trace
	isBot       bool   //if true, act as the bot process instead of the watchdog process
	killOldBot  bool   //if true, search for dangling bot processes and kill them
	watchdogPID int    //stores the watchdog PID if acting as a bot process, used as the exception when killing old bot processes

	//Logging
	log *logger.Logger            //Holds the advanced logger's logging methods
	logPrefix string = "WATCHDOG" //Default to watchdog as it's the initial process
)

func init() {
	//Apply all command-line flags
	flag.StringVar(&cfgPath, "config", "config.json", "the path to the configuration file")
	flag.StringVar(&token, "token", "", "token override for public testing")
	flag.IntVar(&verbosity, "verbosity", 0, "sets the verbosity level; 0 = default, 1 = debug, 2 = trace")
	flag.BoolVar(&isBot, "isBot", false, "act as the bot process instead of the watchdog process")
	flag.BoolVar(&killOldBot, "killOldBot", false, "search for dangling bot processes and kill them")
	flag.IntVar(&watchdogPID, "watchdogPID", -1, "used as the exception when killing old bot processes, requires --killOldBot")
	flag.Parse()

	//Create the logger
	if isBot {
		logPrefix = "BOT" //We're the bot process, report as such
	}
	log = logger.NewLogger(logPrefix, verbosity)
}

func main() {
	log.Trace("--- main() ---")

	if watchdogPID == -1 {
		log.Info("Clinet © JoshuaDoes: 2017-2020.")
	}

	if isBot {
		bot.Bot(cfgPath, token, log)
	} else {
		doWatchdog()
	}

	log.Info("Good-bye!")
}