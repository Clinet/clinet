package main

import (
	"github.com/Clinet/clinet_watchdog"
	"github.com/spf13/pflag"
)

var (
	//Various command-line flags
	featuresFile          string //path to bot feature toggles
	writeFeaturesTemplate bool   //if true, write the current features template to features.template.json
	verbosity             int    //0 = default (info, warning, error), 1 = 0 + debug, 2 = 1 + trace
)

func init() {
	watchdog.Header = "Clinet © JoshuaDoes 2017-2022."
	watchdog.Footer = "Good-bye!"
	watchdog.ImmediateSpawn = true //We want to display Header for a short time
	watchdog.KillOldMain = true //We want to be the only live instance of this bot
}

func main() {
	//Apply all command-line flags
	pflag.StringVar(&featuresFile, "features", "features.json", "path to bot feature toggles")
	pflag.BoolVar(&writeFeaturesTemplate, "writeFeaturesTemplate", true, "write the current features template to features.template.json")
	pflag.IntVar(&verbosity, "verbosity", 0, "sets the verbosity level; 0 = default, 1 = debug, 2 = trace")

	if watchdog.Parse() {
		doBot()
	}
}
