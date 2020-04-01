package bot

import (
	//JSON streaming
	"github.com/Clinet/clinet/utils/json" //JSON wrapper to marshal/unmarshal at will

	//std necessities
	"io/ioutil"
)

type ConfigType int
const (
	ConfigTypeJSON ConfigType = iota
)

type Config struct {
}

//NewConfig creates a new configuration struct with the values in the configuration file
func NewConfig(path string, cfgType ConfigType) (cfg *Config, err error) {
	switch cfgType {
	case ConfigTypeJSON:
		configJSON, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(configJSON, cfg)
	}

	return
}
func (cfg *Config) Load(path string) error {
	return cfg.Load(path)
}