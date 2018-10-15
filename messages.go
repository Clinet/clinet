package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// GuildData holds data specific to a guild
type GuildData struct {
	sync.Mutex //This struct gets accessed very repeatedly throughout various goroutines so we need a mutex to prevent race conditions

	AudioQueue      []AudioQueueEntry
	AudioNowPlaying AudioQueueEntry
	VoiceData       VoiceData
	Queries         map[string]*Query
	YouTubeResults  map[string]*YouTubeResultNav
	SpotifyResults  map[string]*SpotifyResultNav
}

// Query holds data about a query's response message
type Query struct {
	ResponseMessageID string
}

func handleMessage(session *discordgo.Session, message *discordgo.Message, updatedMessageEvent bool) {
	defer recoverPanic()

	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		debugLog("> Error finding message channel", false)
		return //Error finding the channel
	}
	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		debugLog("> Error finding message guild", false)
		return //Error finding the guild
	}
	content := message.Content
	if content == "" {
		return //The message was empty
	}

	contentReplaced, _ := message.ContentWithMoreMentionsReplaced(session)
	eventType := "[New]"
	if updatedMessageEvent {
		eventType = "[Updated]"
	}
	userType := "@"
	if message.Author.Bot {
		userType = "*"
	}
	if strings.Contains(content, "\n") {
		debugLog(eventType+"["+guild.Name+" - #"+channel.Name+"] "+userType+message.Author.Username+"#"+message.Author.Discriminator+":\n"+contentReplaced, false)
	} else {
		debugLog(eventType+"["+guild.Name+" - #"+channel.Name+"] "+userType+message.Author.Username+"#"+message.Author.Discriminator+": "+contentReplaced, false)
	}

	_, guildDataExists := guildData[guild.ID]
	if !guildDataExists {
		guildData[guild.ID] = &GuildData{}
		guildData[guild.ID].VoiceData = VoiceData{}
	}
	guildData[guild.ID].Lock()
	defer guildData[guild.ID].Unlock()

	_, guildSettingsExists := guildSettings[guild.ID]
	if !guildSettingsExists {
		guildSettings[guild.ID] = &GuildSettings{}
	}

	_, userSettingsExists := userSettings[message.Author.ID]
	if !userSettingsExists {
		userSettings[message.Author.ID] = &UserSettings{}
	}

	_, starboardExists := starboards[guild.ID]
	if !starboardExists {
		starboards[guild.ID] = &Starboard{}
		starboards[guild.ID].Emoji = "⭐"
		starboards[guild.ID].NSFWEmoji = "💦"
		starboards[guild.ID].AllowSelfStar = true
		starboards[guild.ID].MinimumStars = 1 //1 for now with testing, default to 2 or 3 later on
	}

	var responseEmbed *discordgo.MessageEmbed

	if guildSettings[guild.ID].BotPrefix != "" {
		if strings.HasPrefix(content, guildSettings[guild.ID].BotPrefix) {
			cmdMsg := strings.TrimPrefix(content, guildSettings[guild.ID].BotPrefix)

			cmd := strings.Split(cmdMsg, " ")

			//0>-1>>>>>-2>>>>>>>>>>>>>>>>>>-3>>>>>>>>>>>>
			//yt search "dance gavin dance" "bloodsucker"
			newCmd := make([]string, 0)
			for i := 0; i < len(cmd); i++ {
				if strings.HasPrefix(cmd[i], "\"") && !strings.HasPrefix(cmd[i], "\"\"") {
					for j := i; j < len(cmd); j++ {
						if strings.HasSuffix(cmd[j], "\"") && !strings.HasSuffix(cmd[j], "\"\"") {
							newArg := strings.Join(cmd[i:j+1], " ")
							newArg = strings.TrimPrefix(newArg, "\"")
							newArg = strings.TrimSuffix(newArg, "\"")
							newCmd = append(newCmd, newArg)
							i = j
							break
						}
					}
				} else {
					newCmd = append(newCmd, cmd[i])
				}
			}
			if len(newCmd) > 0 {
				cmd = newCmd
			}

			member, _ := botData.DiscordSession.GuildMember(guild.ID, message.Author.ID)

			commandEnvironment := &CommandEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Member: member, Command: cmd[0], BotPrefix: guildSettings[guild.ID].BotPrefix, UpdatedMessageEvent: updatedMessageEvent}
			responseEmbed = callCommand(cmd[0], cmd[1:], commandEnvironment)
		}
	} else if strings.HasPrefix(content, botData.CommandPrefix) {
		cmdMsg := strings.TrimPrefix(content, botData.CommandPrefix)

		cmd := strings.Split(cmdMsg, " ")

		//0>-1>>>>>-2>>>>>>>>>>>>>>>>>>-3>>>>>>>>>>>>
		//yt search "dance gavin dance" "bloodsucker"
		newCmd := make([]string, 0)
		for i := 0; i < len(cmd); i++ {
			if strings.HasPrefix(cmd[i], "\"") && !strings.HasPrefix(cmd[i], "\"\"") {
				for j := i; j < len(cmd); j++ {
					if strings.HasSuffix(cmd[j], "\"") && !strings.HasSuffix(cmd[j], "\"\"") {
						newArg := strings.Join(cmd[i:j+1], " ")
						newArg = strings.TrimPrefix(newArg, "\"")
						newArg = strings.TrimSuffix(newArg, "\"")
						newCmd = append(newCmd, newArg)
						i = j
						break
					}
				}
			} else {
				newCmd = append(newCmd, cmd[i])
			}
		}
		if len(newCmd) > 0 {
			cmd = newCmd
		}

		member, _ := botData.DiscordSession.GuildMember(guild.ID, message.Author.ID)

		commandEnvironment := &CommandEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Member: member, Command: cmd[0], BotPrefix: botData.CommandPrefix, UpdatedMessageEvent: updatedMessageEvent}
		responseEmbed = callCommand(cmd[0], cmd[1:], commandEnvironment)
	} else {
		//Swear filter check
		if guildSettings[guild.ID].SwearFilter.Enabled {
			swearFound, swears, err := guildSettings[guild.ID].SwearFilter.Check(content)
			if err != nil {
				//Report error to developer
				ownerPrivChannel, chanErr := session.UserChannelCreate(botData.BotOwnerID)
				if chanErr != nil {
					debugLog("An error occurred creating a private channel with the bot owner.", false)
				} else {
					session.ChannelMessageSend(ownerPrivChannel.ID, "An error occurred with the swear filter: ``"+fmt.Sprintf("%v", err)+"``")
				}
			}
			if swearFound {
				//Log swear event to log channel with list of swears found
				settings, guildFound := guildSettings[guild.ID]
				if guildFound && settings.LogSettings.LoggingEnabled && settings.LogSettings.LoggingEvents.SwearDetect {
					swearDetectEmbed := NewEmbed().
						SetTitle("Logging Event - Swear Detect").
						SetDescription("One or more swears were detected in a message.").
						AddField("Offending User", "<@"+message.Author.ID+">").
						AddField("Source Channel", "<#"+message.ChannelID+">").
						AddField("Swears Detected", strings.Join(swears, ", ")).
						AddField("Offending Message", message.Content).
						InlineAllFields().
						SetColor(0x1C1C1C).MessageEmbed

					session.ChannelMessageSendEmbed(settings.LogSettings.LoggingChannel, swearDetectEmbed)
				}

				//Delete source message
				session.ChannelMessageDelete(message.ChannelID, message.ID)

				//Reply with warning
				msgWarning, _ := session.ChannelMessageSend(message.ChannelID, ":warning: <@"+message.Author.ID+">, please watch your language!")

				//Delete warning after x seconds if x > 0
				if guildSettings[guild.ID].SwearFilter.WarningDeleteTimeout > 0 {
					timer := time.NewTimer(guildSettings[guild.ID].SwearFilter.WarningDeleteTimeout * time.Second)
					<-timer.C
					session.ChannelMessageDelete(msgWarning.ChannelID, msgWarning.ID)
				}

				return
			}
		}

		if botData.BotOptions.UseWolframAlpha || botData.BotOptions.UseDuckDuckGo || botData.BotOptions.UseCustomResponses {
			regexpBotName, _ := regexp.MatchString("^<(@|@\\!)"+session.State.User.ID+">(.*?)$", content) //Ensure prefix is bot tag
			if regexpBotName {
				typingEvent(session, message.ChannelID)

				query := content
				query = strings.Replace(query, "<@!"+session.State.User.ID+">", "", -1)
				query = strings.Replace(query, "<@"+session.State.User.ID+">", "", -1)
				for {
					if strings.HasPrefix(query, " ") {
						query = strings.Replace(query, " ", "", 1)
					} else if strings.HasPrefix(query, ",") {
						query = strings.Replace(query, ",", "", 1)
					} else if strings.HasPrefix(query, ":") {
						query = strings.Replace(query, ":", "", 1)
					} else {
						break
					}
				}

				usedCustomResponse := false
				if botData.BotOptions.UseCustomResponses {
					if len(botData.CustomResponses) > 0 {
						for _, response := range botData.CustomResponses {
							regexpMatched, _ := regexp.MatchString(response.Expression, query)
							if regexpMatched {
								if len(response.CmdResponses) > 0 {
									randomCmd := rand.Intn(len(response.CmdResponses))

									commandEnvironment := &CommandEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Command: response.CmdResponses[randomCmd].CommandName, UpdatedMessageEvent: updatedMessageEvent}
									responseEmbed = callCommand(response.CmdResponses[randomCmd].CommandName, response.CmdResponses[randomCmd].Arguments, commandEnvironment)

									usedCustomResponse = true
								} else if len(response.Responses) > 0 {
									random := rand.Intn(len(response.Responses))

									responseEmbed = response.Responses[random].ResponseEmbed

									usedCustomResponse = true
								}
							}
						}
					}
				}
				if usedCustomResponse == false {
					typingEvent(session, message.ChannelID)

					//Experimental - Use regex for natural language-based commands
					regexCmdPlayComp, err := regexp.Compile(regexCmdPlay)
					if err != nil {
						panic(err)
					}

					matches := regexCmdPlayComp.FindAllString(query, 1) //Get a slice of the results
					if len(matches) > 0 {
						//Remove "Play" from the beginning
						matchSplit := strings.Split(matches[0], " ")
						match := matchSplit[1:]

						commandEnvironment := &CommandEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Command: "play", UpdatedMessageEvent: updatedMessageEvent}
						responseEmbed = commandPlay(match, commandEnvironment)
					} else { //End experimental
						if botData.BotOptions.UseDuckDuckGo {
							responseEmbed, err = queryDuckDuckGo(query)
							if err != nil {
								if botData.BotOptions.UseWolframAlpha {
									responseEmbed, err = queryWolframAlpha(query)
									if err != nil {
										responseEmbed = NewErrorEmbed("Query Error", "We couldn't find the data you were looking for.\nMake sure you're using proper grammar and query structure where applicable.")
									}
								} else {
									responseEmbed = NewErrorEmbed("Query Error", "We couldn't find the data you were looking for.\nMake sure you're using proper grammar and query structure where applicable.")
								}
							}
						} else if botData.BotOptions.UseWolframAlpha {
							responseEmbed, err = queryWolframAlpha(query)
							if err != nil {
								responseEmbed = NewErrorEmbed("Query Error", "We couldn't find the data you were looking for.\nMake sure you're using proper grammar and query structure where applicable.")
							}
						} else {
							responseEmbed = NewErrorEmbed("Query Error", "We couldn't find the data you were looking for.\nMake sure you're using proper grammar and query structure where applicable.")
						}
					}
				}
			}
		}
	}

	if responseEmbed != nil {
		fixedEmbed := Embed{responseEmbed}
		fixedEmbed.Truncate()
		responseEmbed = fixedEmbed.MessageEmbed

		canUpdateMessage := false
		responseID := ""

		_, guildFound := guildData[guild.ID]
		if guildFound {
			if guildData[guild.ID].Queries != nil {
				if guildData[guild.ID].Queries[message.ID] != nil {
					debugLog("> Found previous response", false)
					canUpdateMessage = true
					responseID = guildData[guild.ID].Queries[message.ID].ResponseMessageID
				} else {
					debugLog("> Previous response not found, initializing...", false)
					guildData[guild.ID].Queries[message.ID] = &Query{}
				}
			} else {
				debugLog("> Queries not found, initializing...", false)
				guildData[guild.ID].Queries = make(map[string]*Query)
				debugLog("> Previous response not found, initializing...", false)
				guildData[guild.ID].Queries[message.ID] = &Query{}
			}
		} else {
			debugLog("> Guild not found, initializing...", false)
			guildData[guild.ID] = &GuildData{}
			debugLog("> Queries not found, initializing...", false)
			guildData[guild.ID].Queries = make(map[string]*Query)
			debugLog("> Previous response not found, initializing...", false)
			guildData[guild.ID].Queries[message.ID] = &Query{}
		}

		typingEvent(session, message.ChannelID)

		if canUpdateMessage {
			debugLog("> Editing response...", false)
			session.ChannelMessageEditEmbed(message.ChannelID, responseID, responseEmbed)
		} else {
			debugLog("> Sending response...", false)
			responseMessage, err := session.ChannelMessageSendEmbed(message.ChannelID, responseEmbed)
			if err != nil {
				debugLog("> Error sending response message", false)
			} else {
				debugLog("> Storing response...", false)
				guildData[guild.ID].Queries[message.ID].ResponseMessageID = responseMessage.ID
			}
		}

		stateSave() //Save the state after every interaction
	}
}
