package main

import (
	"os"

	"github.com/JoshuaDoes/logger"
	flag "github.com/spf13/pflag"
)

var (
	//Various command-line flags
	configFile          string //path to bot configuration
	writeConfigTemplate bool   //if true, write the current configuration template to config.template.json
	verbosity           int    //0 = default (info, warning, error), 1 = 0 + debug, 2 = 1 + trace
	isBot               bool   //if true, act as the bot process instead of the watchdog process
	killOldBot          bool   //if true, search for dangling bot processes and kill them
	watchdogPID         int    //stores the watchdog PID if acting as a bot process, used as the exception when killing old bot processes

	//Logging
	log       *logger.Logger
	logPrefix string = "watchdog"
)

func init() {
	//Apply all command-line flags
	flag.StringVar(&configFile, "config", "config.json", "path to bot configuration")
	flag.BoolVar(&writeConfigTemplate, "writeConfigTemplate", true, "write the current configuration template to config.template.json")
	flag.IntVar(&verbosity, "verbosity", 0, "sets the verbosity level; 0 = default, 1 = debug, 2 = trace")
	flag.BoolVar(&isBot, "isBot", false, "act as the bot process instead of the watchdog process")
	flag.BoolVar(&killOldBot, "killOldBot", false, "search for dangling bot processes and kill them")
	flag.IntVar(&watchdogPID, "watchdogPID", -1, "used as the exception when killing old bot processes, requires --killOldBot")
	flag.Parse()

	//Create the logger
	if isBot {
		logPrefix = "bot" //We're the bot process, report as such
	}
	log = logger.NewLogger(logPrefix, verbosity)
}

func main() {
	log.Trace("--- main() ---")

	if watchdogPID == -1 {
		log.Info("Clinet © JoshuaDoes 2017-2022.")
	}

	if isBot {
		doBot()
		os.Exit(0)
	}

	doWatchdog()
	log.Info("Good-bye!")
}
