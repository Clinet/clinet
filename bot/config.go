package bot

import (
	//JSON streaming
	"github.com/Clinet/clinet/utils/json" //JSON wrapper to marshal/unmarshal at will

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
}

//NewConfig creates a new configuration struct with the values in the configuration file
func NewConfig(path string, cfgType ConfigType) (cfg *Config, err error) {
	Log.Trace("--- NewConfig(", path, ", ", cfgType, ") ---")

	cfg = &Config{}

	switch cfgType {
	case ConfigTypeJSON:
		configJSON, err := ioutil.ReadFile(path)
		if err != nil {
			Log.Error("Error reading configuration file")
			return nil, err
		}

		err = json.Unmarshal(configJSON, cfg)
	default:
		Log.Error("Unknown configuration type ", cfgType)
		return nil, errors.New("bot: config: unknown configuration type")
	}

	return
}
func (cfg *Config) Load(path string) error {
	return cfg.Load(path)
}

type CfgDiscord struct {
	//Stuff for communication with Discord
	Token string `json:"token"`

	//Trust for Discord communication
	DisplayName   string `json:"displayName"`   //The display name for communicating on Discord
	OwnerID       string `json:"ownerID"`       //The user ID of the bot owner on Discord
	CommandPrefix string `json:"commandPrefix"` //The command prefix to use when invoking the bot on Discord
}