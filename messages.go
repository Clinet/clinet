package main

import (
	//	"bytes"
	"fmt"
	//	"image"
	//	"image/png"
	"math/rand"
	//	"net/http"
	//	"net/url"
	"regexp"
	//	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	//	"github.com/disintegration/gift"
	//	"github.com/rylio/ytdl"
)

// Query holds data about a query's response message
type Query struct {
	ResponseMessageID string
}

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	defer recoverPanic()

	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: "+fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}

	go handleMessage(session, message, false)
}
func discordMessageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	defer recoverPanic()

	message, err := session.ChannelMessage(event.ChannelID, event.ID) //Make it easier to keep track of what's happening
	if err != nil {
		debugLog("> Error fnding message: "+fmt.Sprintf("%v", err), false)
		return //Error finding message
	}
	if message.Author.ID == session.State.User.ID {
		debugLog("> Message author ID matched bot ID, ignoring message", false)
		return //The bot should never reply to itself
	}

	go handleMessage(session, message, true)
}
func discordMessageDelete(session *discordgo.Session, event *discordgo.MessageDelete) {
	defer recoverPanic()

	message := event //Make it easier to keep track of what's happening

	debugLog("[D] ID: "+message.ID, false)

	guildChannel, err := session.Channel(message.ChannelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			_, messageFound := guildData[guildID].Queries[message.ID]
			if messageFound {
				debugLog("> Deleting message...", false)
				session.ChannelMessageDelete(message.ChannelID, guildData[guildID].Queries[message.ID].ResponseMessageID) //Delete the query response message
				guildData[guildID].Queries[message.ID] = nil                                                              //Remove the message from the query list
			} else {
				debugLog("> Error finding deleted message in queries list", false)
			}
		} else {
			debugLog("> Error finding guild for deleted message", false)
		}
	} else {
		debugLog("> Error finding channel for deleted message", false)
	}
}
func discordMessageDeleteBulk(session *discordgo.Session, event *discordgo.MessageDeleteBulk) {
	defer recoverPanic()

	messages := event.Messages
	channelID := event.ChannelID

	guildChannel, err := session.Channel(channelID)
	if err == nil {
		guildID := guildChannel.GuildID

		_, guildFound := guildData[guildID]
		if guildFound {
			for i := 0; i > len(messages); i++ {
				debugLog("[D] ID: "+messages[i], false)
				_, messageFound := guildData[guildID].Queries[messages[i]]
				if messageFound {
					debugLog("> Deleting message...", false)
					session.ChannelMessageDelete(channelID, guildData[guildID].Queries[messages[i]].ResponseMessageID) //Delete the query response message
					guildData[guildID].Queries[messages[i]] = nil                                                      //Remove the message from the query list
				} else {
					debugLog("> Error finding deleted message in queries list", false)
				}
			}
		}
	}
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

	/*
		//If message is single-lined
			[New][District JD - #main] @JoshuaDoes#0001: Hello, world!

		//If message is multi-lined
			[New][District JD - #main] @JoshuaDoes#0001:
			Hello, world!
			My name is Joshua.
			This is a lot of fun!

		//If user is bot
			[New][District JD - #main] *Clinet#1823: Hello, world!
	*/
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

	var responseEmbed *discordgo.MessageEmbed

	if guildData[guild.ID] == nil {
		guildData[guild.ID] = &GuildData{}
		guildData[guild.ID].VoiceData = VoiceData{}
	}

	if guildSettings[guild.ID] == nil {
		guildSettings[guild.ID] = &GuildSettings{}
	}

	if userSettings[message.Author.ID] == nil {
		userSettings[message.Author.ID] = &UserSettings{}
	}

	if strings.HasPrefix(content, botData.CommandPrefix) {
		cmdMsg := strings.Replace(content, botData.CommandPrefix, "", 1)
		cmd := strings.Split(cmdMsg, " ")

		commandEnvironment := &CommandEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Command: cmd[0], UpdatedMessageEvent: updatedMessageEvent}
		responseEmbed = callCommand(cmd[0], cmd[1:], commandEnvironment)
	} else {
		if botData.BotOptions.UseWolframAlpha || botData.BotOptions.UseDuckDuckGo || botData.BotOptions.UseCustomResponses {
			//regexpBotName, _ := regexp.MatchString("<(@|@\\!)"+session.State.User.ID+">(.*?)", content)
			regexpBotName, _ := regexp.MatchString("^<(@|@\\!)"+session.State.User.ID+">(.*?)$", content) //Ensure prefix is bot tag
			//if regexpBotName && strings.HasSuffix(content, "?") {
			if regexpBotName { //Experiment with not requiring question mark suffix
				if !updatedMessageEvent {
					typingEvent(session, message.ChannelID)
				}
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
									break
								} else if len(response.Responses) > 0 {
									random := rand.Intn(len(response.Responses))

									responseEmbed = response.Responses[random].ResponseEmbed

									usedCustomResponse = true
									break
								}
							}
						}
					}
				}
				if usedCustomResponse == false {
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
										responseEmbed = NewErrorEmbed("Query Error", "No response was found.")
									}
								} else {
									responseEmbed = NewErrorEmbed("Query Error", "No response was found.")
								}
							}
						} else if botData.BotOptions.UseWolframAlpha {
							responseEmbed, err = queryWolframAlpha(query)
							if err != nil {
								responseEmbed = NewErrorEmbed("Query Error", "No response was found.")
							}
						} else {
							responseEmbed = NewErrorEmbed("Query Error", "No response was found.")
						}
					}
				}
			}
		}
	}

	if responseEmbed != nil {
		if !updatedMessageEvent {
			typingEvent(session, message.ChannelID)
		}

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
	}
}
