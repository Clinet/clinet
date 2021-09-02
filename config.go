package main

import (
	//JSON streaming
	"github.com/JoshuaDoes/json" //JSON wrapper to marshal/unmarshal at will

	//std necessities
	"errors"
	"io/ioutil"
)

type ConfigType int
const (
	ConfigTypeJSON ConfigType = iota
	ConfigTypeTOML
	ConfigTypeXML
)

type Config struct {
	Discord CfgDiscord `json:"discord"`

	path string //The path to the configuration file
}

//loadConfig creates a new configuration struct with the values in the specified configuration file
func loadConfig(path string, cfgType ConfigType) (cfg *Config, err error) {
	log.Trace("--- loadConfig(", path, ", ", cfgType, ") ---")

	cfg = &Config{path: path}

	switch cfgType {
	case ConfigTypeJSON:
		configJSON, err := ioutil.ReadFile(path)
		if err != nil {
			log.Error("Error reading configuration file:", err)
			return nil, err
		}

		err = json.Unmarshal(configJSON, cfg)
	default:
		log.Error("Unknown configuration type:", cfgType)
		return nil, errors.New("bot: config: unknown configuration type")
	}

	return
}

func saveConfig(cfg *Config, path string, cfgType ConfigType) (err error) {
	log.Trace("--- saveConfig(", path, ", ", cfgType, ") ---")

	configJSON, err := json.Marshal(cfg, true)
	if err != nil {
		log.Error("Error generating config JSON:", err)
		return err
	}

	err = ioutil.WriteFile(path, configJSON, 0644)
	if err != nil {
		log.Error("Error saving config JSON to path:", err)
	}
	return err
}

//LoadFrom loads the configuration from the specified path into the current cfg
func (cfg *Config) LoadFrom(path string, cfgType ConfigType) (err error) {
	cfg, err = loadConfig(path, cfgType)
	return err
}
//SaveTo saves the current cfg to the specified path
func (cfg *Config) SaveTo(path string, cfgType ConfigType) (err error) {
	err = saveConfig(cfg, path, cfgType)
	return err
}