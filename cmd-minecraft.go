package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"4d63.com/tz"
	"github.com/Syfaro/minepong"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/minotar/minecraft"
)

type MCServerDescription struct {
	Extra []MCServerDescriptionExtra `json:"extra"`
	Text  string                     `json:"text"`
}
type MCServerDescriptionExtra struct {
	Text          string `json:"text"`
	Color         string `json:"color"`
	Obfuscated    bool   `json:"obfuscated"`
	Bold          bool   `json:"bold"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Italic        bool   `json:"italic"`
}
type MCOriginalUser struct {
	ID       string `json:"id"`
	Username string `json:"name"`
}
type MCUsernameHistory struct {
	Username    string `json:"name"`
	TimeChanged int64  `json:"changedToAt"`
}

func GetAPIOldProfile(mc *minecraft.Minecraft, username string) (*minecraft.APIProfileResponse, error) {
	originalUser := &MCOriginalUser{}

	url := fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s?at=0", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", mc.UserAgent)

	res, err := mc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	err = unmarshal(res, originalUser)
	if err != nil {
		return nil, err
	}

	if originalUser.ID == "" || originalUser.Username == "" {
		return nil, errors.New("not found")
	}

	profile, err := mc.GetAPIProfile(originalUser.Username)
	return &profile, err
}
func GetAPIUsernameHistory(mc *minecraft.Minecraft, uuid string) ([]MCUsernameHistory, error) {
	history := make([]MCUsernameHistory, 0)

	url := fmt.Sprintf("https://api.mojang.com/user/profiles/%s/names", uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", mc.UserAgent)

	res, err := mc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	err = unmarshal(res, &history)
	if err != nil {
		return nil, err
	}

	return history, nil
}

func commandMinecraft(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	timezone := userSettings[env.User.ID].Timezone
	if timezone == "" {
		return NewErrorEmbed("Minecraft Error", "Please set a timezone first!\n\nEx: ``"+env.BotPrefix+"user timezone America/New_York``")
	}
	location, err := tz.LoadLocation(timezone)
	if err != nil {
		return NewErrorEmbed("Minecraft Error", "You have an invalid timezone set, please set a new one first!\n\nEx: ``"+env.BotPrefix+"user timezone America/New_York``")
	}

	switch args[0] {
	case "user", "player", "avatar", "skin", "uuid":
		minecraftAPI := minecraft.NewMinecraft()

		profileAPI, err := minecraftAPI.GetAPIProfile(args[1])
		if err != nil {
			oldProfileAPI, err := GetAPIOldProfile(minecraftAPI, args[1])
			if err != nil {
				return NewErrorEmbed("Minecraft Error", "Invalid or unknown username ``"+args[1]+"``.")
			}
			profileAPI = *oldProfileAPI
		}

		minecraftEmbed := NewEmbed().
			SetTitle("Minecraft - " + profileAPI.User.Username).
			SetImage("https://minotar.net/armor/body/" + profileAPI.User.UUID + ".png").
			SetFooter("UUID: " + profileAPI.User.UUID).
			SetColor(0x5B8731)

		if profileAPI.Legacy {
			minecraftEmbed.AddField("Account Type", "Legacy")
		} else if profileAPI.Demo {
			minecraftEmbed.AddField("Account Type", "Demo")
		}

		usernameHistory, _ := GetAPIUsernameHistory(minecraftAPI, profileAPI.User.UUID)
		if len(usernameHistory) > 0 {
			history := make([]string, 0)
			for _, username := range usernameHistory {
				if username.TimeChanged != 0 {
					timeChanged := time.Unix(username.TimeChanged, 0).In(location)
					history = append(history, username.Username+" (changed "+humanize.Time(timeChanged.In(location))+" at ``"+timeChanged.In(location).String()+"``)")
				} else {
					history = append(history, username.Username)
				}
			}
			if len(history) > 0 {
				minecraftEmbed.AddField("Username History", strings.Join(history, "\n"))
			}
		}

		profileSession, err := minecraftAPI.GetSessionProfile(profileAPI.User.UUID)
		if err == nil {
			for _, property := range profileSession.Properties {
				switch property.Name {
				case "textures":
					continue
				}
				minecraftEmbed.AddField("Property: "+property.Name, property.Value)
			}
		}

		return minecraftEmbed.MessageEmbed
	case "server", "host", "ip", "ping":
		host := args[1]
		if ipPort := strings.Split(args[1], ":"); len(ipPort) == 1 {
			host += ":25565"
		}

		server, err := minepong.Ping(host)
		if err != nil {
			return NewErrorEmbed("Minecraft Error", "Invalid or unknown server ``"+args[1]+"``.")
		}

		title := "Minecraft - " + args[1]
		if server.ResolvedHost != "" {
			title = "Minecraft - " + server.ResolvedHost
		}

		version := strconv.Itoa(server.Version.Protocol)
		if server.Version.Name != "" {
			version = server.Version.Name + "\n" + version
		}

		minecraftEmbed := NewEmbed().
			SetTitle(title).
			AddField("Version", version).
			AddField("Players", strconv.Itoa(server.Players.Online)+"/"+strconv.Itoa(server.Players.Max)).
			SetColor(0x5B8731)

		if server.FavIcon != "" {
			hostReplaced := strings.Replace(host, ":", "/", 1)
			minecraftEmbed.SetThumbnail("https://api.minetools.eu/favicon/" + hostReplaced)
		}

		switch v := server.Description.(type) {
		case map[string]interface{}:
			jsonData, err := json.Marshal(server.Description)
			if err == nil {
				var advancedDescription *MCServerDescription
				if err := json.Unmarshal(jsonData, &advancedDescription); err == nil {
					if len(advancedDescription.Extra) > 0 {
						formattedDescription := ""
						for _, extra := range advancedDescription.Extra {
							text := strings.TrimPrefix(strings.TrimSuffix(extra.Text, " "), " ")
							if extra.Bold {
								text = "**" + text + "**"
							}
							if extra.Strikethrough {
								text = "~~" + text + "~~"
							}
							if extra.Underline {
								text = "__" + text + "__"
							}
							if extra.Italic {
								text = "*" + text + "*"
							}
							formattedDescription += text + " "
						}
						minecraftEmbed.AddField("Description", formattedDescription)
					} else if advancedDescription.Text != "" {
						minecraftEmbed.AddField("Description", advancedDescription.Text)
					}
				}
			}
		case string:
			if v != "" {
				format := false
				obfuscated := false
				bold := false
				strikethrough := false
				underline := false
				italic := false

				formattedDescription := ""
				for _, character := range v {
					switch string(character) {
					case "§":
						format = true
						continue
					case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f":
						if format {
							format = false
							continue
						}
					case "k":
						if format {
							format = false
							obfuscated = true
							continue
						}
					case "l":
						if format {
							format = false
							if bold == false {
								bold = true
								formattedDescription += "**"
							}
							continue
						}
					case "m":
						if format {
							format = false
							if strikethrough == false {
								strikethrough = true
								formattedDescription += "~~"
							}
							continue
						}
					case "n":
						if format {
							format = false
							if underline == false {
								underline = true
								formattedDescription += "__"
							}
							continue
						}
					case "o":
						if format {
							format = false
							if italic == false {
								italic = true
								formattedDescription += "*"
							}
							continue
						}
					case "r":
						if format {
							format = false
							obfuscated = false
							if bold {
								bold = false
								formattedDescription += "**"
							}
							if strikethrough {
								strikethrough = false
								formattedDescription += "~~"
							}
							if underline {
								underline = false
								formattedDescription += "__"
							}
							if italic {
								italic = false
								formattedDescription += "*"
							}
							continue
						}
					}
					if obfuscated {
						formattedDescription += "▓"
					} else {
						formattedDescription += string(character)
					}
				}

				if bold {
					formattedDescription += "**"
				}
				if strikethrough {
					formattedDescription += "~~"
				}
				if underline {
					formattedDescription += "__"
				}
				if italic {
					formattedDescription += "*"
				}

				minecraftEmbed.AddField("Description", formattedDescription)
			}
		default:
			minecraftEmbed.AddField("Unknown Type", fmt.Sprintf("%T", v))
		}

		return minecraftEmbed.MessageEmbed
	}

	return NewErrorEmbed("Minecraft Error", "Unknown command ``"+args[1]+"``.")
}
