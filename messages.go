package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/JoshuaDoes/go-wolfram"
	"github.com/bwmarrin/discordgo"
)

// Query holds data about a query's response message
type Query struct {
	ResponseMessageID string
}

func debugMessage(session *discordgo.Session, message *discordgo.Message, channel *discordgo.Channel, guild *discordgo.Guild, updatedMessageEvent bool) {
	content := message.Content
	if content == "" {
		if len(message.Embeds) > 0 {
			for _, embed := range message.Embeds {
				debugEmbed(embed, message.Author, channel, guild, updatedMessageEvent)
			}
		}
		return //The message was empty
	}
	contentReplaced, err := message.ContentWithMoreMentionsReplaced(session)
	if err != nil {
		contentReplaced = content
	}
	eventType := "New"
	if updatedMessageEvent {
		eventType = "Updated"
	}
	userType := "@"
	if message.Author.Bot {
		userType = "*"
	}

	if strings.Contains(content, "\n") {
		Debug.Printf("[%s][%s - #%s] %s%s#%s:\n%s", eventType, guild.Name, channel.Name, userType, message.Author.Username, message.Author.Discriminator, contentReplaced)
	} else {
		Debug.Printf("[%s][%s - #%s] %s%s#%s: %s", eventType, guild.Name, channel.Name, userType, message.Author.Username, message.Author.Discriminator, contentReplaced)
	}
}

func debugEmbed(embed *discordgo.MessageEmbed, author *discordgo.User, channel *discordgo.Channel, guild *discordgo.Guild, updatedMessageEvent bool) {
	embedJSON, err := json.MarshalIndent(embed, "", "\t")
	if err != nil {
		return
	}
	eventType := "New"
	if updatedMessageEvent {
		eventType = "Updated"
	}
	userType := "@"
	if author.Bot {
		userType = "*"
	}

	Debug.Printf("[%s][%s - #%s] %s%s#%s:\n%s", eventType, guild.Name, channel.Name, userType, author.Username, author.Discriminator, string(embedJSON))
}

func handleMessage(session *discordgo.Session, message *discordgo.Message, updatedMessageEvent bool) {
	defer recoverPanic()

	if message.Author.Bot {
		return //We don't want bots to interact with our bot
	}

	channel, err := session.State.Channel(message.ChannelID)
	if err != nil {
		return //Error finding the channel
	}
	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		return //Error finding the guild
	}
	content := message.Content
	if content == "" {
		return //The message was empty
	}
	member, err := botData.DiscordSession.GuildMember(guild.ID, message.Author.ID)
	if err != nil {
		return //Error finding the guild member
	}

	//Initialize various datapoints
	initializeGuildData(guild.ID)
	initializeGuildSettings(guild.ID)
	initializeUserSettings(message.Author.ID)
	initializeStarboard(guild.ID)

	guildData[guild.ID].Lock()
	defer guildData[guild.ID].Unlock()

	//The embed that will be sent off to Discord
	var responseEmbed *discordgo.MessageEmbed

	for _, roleMe := range guildSettings[guild.ID].RoleMeList {
		for _, trigger := range roleMe.Triggers {
			if roleMe.CaseSensitive {
				if trigger == content {
					handleRoleMe(roleMe, guild.ID, channel.ID, message.Author.ID)
					break
				}
			} else {
				if strings.EqualFold(trigger, content) {
					handleRoleMe(roleMe, guild.ID, channel.ID, message.Author.ID)
					break
				}
			}
		}
	}

	regexpBotName, _ := regexp.MatchString("^<(@|@\\!)"+session.State.User.ID+">(.*?)$", content) //Ensure prefix is bot tag
	prefix := ""
	if guildSettings[guild.ID].BotPrefix != "" {
		if strings.HasPrefix(content, guildSettings[guild.ID].BotPrefix) {
			prefix = guildSettings[guild.ID].BotPrefix
		}
	} else {
		if strings.HasPrefix(content, botData.CommandPrefix) {
			prefix = botData.CommandPrefix
		}
	}

	if regexpBotName {
		if botData.BotOptions.UseWolframAlpha || botData.BotOptions.UseDuckDuckGo || botData.BotOptions.UseCustomResponses {
			debugMessage(session, message, channel, guild, updatedMessageEvent)
			typingEvent(session, message.ChannelID, updatedMessageEvent)

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

			commandEnvironment := &CommandEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Member: member, UpdatedMessageEvent: updatedMessageEvent}
			responseEmbed = callNLP(query, commandEnvironment)

			if responseEmbed == nil {
				typingEvent(session, message.ChannelID, updatedMessageEvent)

				var previousConversation *wolfram.Conversation

				if guildData[guild.ID].WolframConversations != nil {
					if guildData[guild.ID].WolframConversations[message.Author.ID] != nil {
						previousConversation = guildData[guild.ID].WolframConversations[message.Author.ID]
					} else {
						guildData[guild.ID].WolframConversations[message.Author.ID] = &wolfram.Conversation{}
					}
				} else {
					guildData[guild.ID].WolframConversations = make(map[string]*wolfram.Conversation)
					guildData[guild.ID].WolframConversations[message.Author.ID] = &wolfram.Conversation{}
				}

				queryEnvironment := &QueryEnvironment{Channel: channel, Guild: guild, Message: message, User: message.Author, Member: member, WolframConversation: previousConversation}
				responseEmbed, err = getQueryResult(query, queryEnvironment)
				if err != nil {
					responseEmbed = NewErrorEmbed("Query Error", "We couldn't find a service to handle your query.\nMake sure you're using proper grammar and query structure where applicable.")
				}
			}
		}
	} else if prefix != "" {
		debugMessage(session, message, channel, guild, updatedMessageEvent)

		cmdMsg := strings.TrimPrefix(content, prefix)

		cmd := strings.Split(cmdMsg, " ")

		//0>>>>>>-1>>>>>-2>>>>>>>>>>>>>>>>>>-3>>>>>>>>>>
		//spotify search "dance gavin dance" bloodsucker
		//0: spotify
		//1: search
		//2: dance gavin dance
		//3: bloodsucker
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
	}

	//Swear filter check
	if guildSettings[guild.ID].SwearFilter.Enabled && responseEmbed == nil {
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

	if responseEmbed == InternalEmbedActionCompleted {
		return
	}

	if responseEmbed != nil {
		fixedEmbed := Embed{responseEmbed}
		fixedEmbed.Truncate()
		responseEmbed = fixedEmbed.MessageEmbed

		canUpdateMessage := false
		responseID := ""

		if guildData[guild.ID].Queries != nil {
			if guildData[guild.ID].Queries[message.ID] != nil {
				canUpdateMessage = true
				responseID = guildData[guild.ID].Queries[message.ID].ResponseMessageID
			} else {
				guildData[guild.ID].Queries[message.ID] = &Query{}
			}
		} else {
			guildData[guild.ID].Queries = make(map[string]*Query)
			guildData[guild.ID].Queries[message.ID] = &Query{}
		}

		if canUpdateMessage {
			session.ChannelMessageEditEmbed(message.ChannelID, responseID, responseEmbed)
			debugEmbed(responseEmbed, botData.DiscordSession.State.User, channel, guild, updatedMessageEvent)
		} else {
			typingEvent(session, message.ChannelID, updatedMessageEvent)

			responseMessage, err := session.ChannelMessageSendEmbed(message.ChannelID, responseEmbed)
			if err == nil {
				debugEmbed(responseEmbed, botData.DiscordSession.State.User, channel, guild, updatedMessageEvent)
				guildData[guild.ID].Queries[message.ID].ResponseMessageID = responseMessage.ID
			}
		}

		stateSave() //Save the state after every interaction
	}
}
