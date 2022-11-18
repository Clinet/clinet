package main

import (
	"github.com/Clinet/clinet_watchdog"
	"github.com/spf13/pflag"
)

var (
	//Various command-line flags
	configFile          string //path to bot configuration
	writeConfigTemplate bool   //if true, write the current configuration template to config.template.json
	verbosity           int    //0 = default (info, warning, error), 1 = 0 + debug, 2 = 1 + trace
)

func init() {
	watchdog.Header = "Clinet © JoshuaDoes 2017-2022."
	watchdog.Footer = "Good-bye!"
}

func main() {
	//Apply all command-line flags
	pflag.StringVar(&configFile, "config", "config.json", "path to bot configuration")
	pflag.BoolVar(&writeConfigTemplate, "writeConfigTemplate", true, "write the current configuration template to config.template.json")
	pflag.IntVar(&verbosity, "verbosity", 0, "sets the verbosity level; 0 = default, 1 = debug, 2 = trace")

	if watchdog.Parse() {
		doBot()
	}
}
