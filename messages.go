package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rylio/ytdl"
)

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

		switch cmd[0] {
		/*		case "balance":
					if len(cmd) > 1 {
						newBalance, err := strconv.Atoi(cmd[1])
						if err != nil {
							responseEmbed = NewEmbed().
								SetTitle("Balance Error").
								SetDescription("Invalid number.").
								SetColor(0x1C1C1C).MessageEmbed
						} else {
							userSettings[message.Author.ID].Balance = newBalance
							responseEmbed = NewEmbed().
								SetTitle("Balance Updated").
								SetDescription("Your new balance is " + strconv.Itoa(newBalance) + " clies.").
								SetColor(0x85BB65).MessageEmbed
						}
					} else {
						responseEmbed = NewEmbed().
							SetTitle("Balance Receipt").
							SetDescription("Your current balance is " + strconv.Itoa(userSettings[message.Author.ID].Balance) + " clies.").
							SetColor(0x85BB65).MessageEmbed
					}
				case "description", "desc", "aboutme":
					if len(cmd) > 1 {
						userSettings[message.Author.ID].Description = strings.Join(cmd[1:], " ")
						responseEmbed = NewEmbed().
							SetTitle("About - " + message.Author.Username + " - Updated").
							SetDescription(userSettings[message.Author.ID].Description).
							SetColor(0x1C1C1C).MessageEmbed
					} else {
						responseEmbed = NewEmbed().
							SetTitle("About - " + message.Author.Username).
							SetDescription(userSettings[message.Author.ID].Description).
							SetColor(0x1C1C1C).MessageEmbed
					} */
		case "help":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - Help").
				SetDescription("A list of available commands for "+botData.BotName+".").
				AddField(botData.CommandPrefix+"help", "Displays this help message.").
				AddField(botData.CommandPrefix+"about", "Displays information about "+botData.BotName+" and how to use it.").
				AddField(botData.CommandPrefix+"version", "Displays the current version of "+botData.BotName+".").
				AddField(botData.CommandPrefix+"credits", "Displays a list of credits for the creation and functionality of "+botData.BotName+".").
				AddField(botData.CommandPrefix+"roll", "Rolls a dice.").
				AddField(botData.CommandPrefix+"doubleroll", "Rolls two die.").
				AddField(botData.CommandPrefix+"coinflip", "Flips a coin.").
				AddField(botData.CommandPrefix+"xkcd (comic number|random|latest)", "Displays an xkcd comic depending on the requested type or comic number.").
				AddField(botData.CommandPrefix+"imgur (url)", "Displays info about the specified Imgur image, album, gallery image, or gallery album.").
				AddField(botData.CommandPrefix+"github/gh username(/repo_name)", "Displays info about the specified GitHub user or repo.").
				AddField(botData.CommandPrefix+"play (YouTube search query, YouTube URL, SoundCloud URL, or direct audio/video URL (as supported by ffmpeg))", "Plays either the first result from a YouTube search query or the specified stream URL in the user's current voice channel.").
				AddField(botData.CommandPrefix+"youtube search (query)", "Displays paginated results of the specified YouTube search query with a command list for navigating and selecting a result.").
				AddField(botData.CommandPrefix+"pause", "If already playing, pauses the current audio stream.").
				AddField(botData.CommandPrefix+"resume", "If previously paused, resumes the current audio stream.").
				AddField(botData.CommandPrefix+"stop", "Stops the currently playing audio.").
				AddField(botData.CommandPrefix+"skip", "Stops the currently playing audio, and, if available, attempts to play the next audio in the queue.").
				AddField(botData.CommandPrefix+"repeat", "Switches the repeat level between the entire guild queue, the currently now playing audio, and not repeating at all.").
				AddField(botData.CommandPrefix+"shuffle", "Shuffles the current guild queue.").
				AddField(botData.CommandPrefix+"queue help", "Lists all available queue commands.").
				AddField(botData.CommandPrefix+"nowplaying/np", "Gets info about the currently playing audio.").
				AddField(botData.CommandPrefix+"leave", "Leaves the current voice channel.").
				SetColor(0xFAFAFA).MessageEmbed
		case "about":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - About").
				SetDescription(botData.BotName+" is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
				AddField("How can I use "+botData.BotName+" in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
				AddField("How can I help keep "+botData.BotName+" running?", "The best ways to help keep "+botData.BotName+" running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
				AddField("How can I use "+botData.BotName+"?", "There are many ways to make use of "+botData.BotName+".\n1) Type ``cli$help`` and try using some of the available commands.\n2) Ask "+botData.BotName+" a question, ex: ``@"+botData.BotName+"#1823, what time is it?`` or ``@"+botData.BotName+"#1823, what is DiscordApp?``.").
				AddField("Where can I join the "+botData.BotName+" Discord server?", "If you would like to get help and support with "+botData.BotName+" or experiment with the latest and greatest of "+botData.BotName+", use the Discord Server Invite Link at the end of this message.").
				AddField("Invite Link", "https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot").
				AddField("Donation Link", "https://www.paypal.me/JoshuaDoes").
				AddField("Source Code Link", "https://github.com/JoshuaDoes/clinet-discord/").
				AddField("Discord Server Invite Link", "https://discord.gg/qkbKEWT").
				SetColor(0x1C1C1C).MessageEmbed
		case "version":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - Version").
				AddField("Build ID", BuildID).
				AddField("Build Date", BuildDate).
				AddField("Latest Development", GitCommitMsg).
				AddField("GitHub Commit URL", GitHubCommitURL).
				AddField("Golang Version", GolangVersion).
				SetColor(0x1C1C1C).MessageEmbed
		case "credits":
			responseEmbed = NewEmbed().
				SetTitle(botData.BotName+" - Credits").
				AddField("Bot Development", "- JoshuaDoes (2018)").
				AddField("Programming Language", "- Golang").
				AddField("Golang Libraries", "- https://github.com/bwmarrin/discordgo\n"+
					"- https://github.com/JoshuaDoes/duckduckgolang\n"+
					"- https://github.com/google/go-github/github\n"+
					"- https://github.com/jonas747/dca\n"+
					"- https://github.com/JoshuaDoes/go-soundcloud\n"+
					"- https://github.com/JoshuaDoes/go-wolfram\n"+
					"- https://github.com/koffeinsource/go-imgur\n"+
					"- https://github.com/koffeinsource/go-klogger\n"+
					"- https://github.com/nishanths/go-xkcd\n"+
					"- https://github.com/paked/configure\n"+
					"- https://github.com/robfig/cron\n"+
					"- https://github.com/rylio/ytdl\n"+
					"- https://google.golang.org/api/googleapi/transport\n"+
					"- https://google.golang.org/api/youtube/v3").
				AddField("Icon Design", "- thejsa").
				AddField("Source Code", "- https://github.com/JoshuaDoes/clinet-discord").
				SetColor(0x1C1C1C).MessageEmbed
		case "roll":
			random := rand.Intn(6) + 1
			responseEmbed = NewGenericEmbed("Roll", "You rolled a "+strconv.Itoa(random)+"!")
		case "doubleroll", "rolldouble":
			random1 := rand.Intn(6) + 1
			random2 := rand.Intn(6) + 1
			randomTotal := random1 + random2
			responseEmbed = NewGenericEmbed("Double Roll", "You rolled a "+strconv.Itoa(random1)+" and a "+strconv.Itoa(random2)+". The total is "+strconv.Itoa(randomTotal)+"!")
		case "coinflip", "flipcoin":
			random := rand.Intn(2)
			switch random {
			case 0:
				responseEmbed = NewGenericEmbed("Coin Flip", "You got heads!")
			case 1:
				responseEmbed = NewGenericEmbed("Coin Flip", "You got tails!")
			}
		case "imgur":
			if len(cmd) > 1 {
				responseEmbed, err = queryImgur(cmd[1])
				if err != nil {
					responseEmbed = NewErrorEmbed("Imgur Error", fmt.Sprintf("%v", err))
				}
			} else {
				responseEmbed = NewErrorEmbed("Imgur Error", "You must specify an Imgur URL to query Imgur with.")
			}
		case "xkcd":
			if len(cmd) > 1 {
				switch cmd[1] {
				case "random": //Get random XKCD comic
					comic, err := botData.BotClients.XKCD.Random()
					if err != nil {
						responseEmbed = NewErrorEmbed("XKCD Error", "Could not find a random XKCD comic.")
					} else {
						responseEmbed = NewEmbed().
							SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
							SetDescription(comic.Title).
							SetImage(comic.ImageURL).
							SetColor(0x96A8C8).MessageEmbed
					}
				case "latest": //Get latest XKCD comic
					comic, err := botData.BotClients.XKCD.Latest()
					if err != nil {
						responseEmbed = NewErrorEmbed("XKCD Error", "Could not find the latest XKCD comic.")
					} else {
						responseEmbed = NewEmbed().
							SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
							SetDescription(comic.Title).
							SetImage(comic.ImageURL).
							SetColor(0x96A8C8).MessageEmbed
					}
				default: //Get specified XKCD comic
					comicNumber, err := strconv.Atoi(cmd[1])
					if err != nil { //Specified comic is not a valid integer
						responseEmbed = NewErrorEmbed("XKCD Error", "``"+cmd[1]+"`` is not a valid number.")
					} else {
						comic, err := botData.BotClients.XKCD.Get(comicNumber)
						if err != nil {
							responseEmbed = NewErrorEmbed("XKCD Error", "Could not find XKCD comic #"+cmd[1]+".")
						} else {
							responseEmbed = NewEmbed().
								SetTitle("xkcd - #" + cmd[1]).
								SetDescription(comic.Title).
								SetImage(comic.ImageURL).
								SetColor(0x96A8C8).MessageEmbed
						}
					}
				}
			} else { //Get random XKCD comic
				comic, err := botData.BotClients.XKCD.Random()
				if err != nil {
					responseEmbed = NewErrorEmbed("XKCD Error", "Error finding random XKCD comic.")
				} else {
					responseEmbed = NewEmbed().
						SetTitle("xkcd - #" + strconv.Itoa(comic.Number)).
						SetDescription(comic.Title).
						SetImage(comic.ImageURL).
						SetColor(0x96A8C8).MessageEmbed
				}
			}
		case "github", "gh":
			// https://godoc.org/github.com/google/go-github/github
			if len(cmd) > 1 {
				request := strings.Split(cmd[1], "/")
				switch len(request) {
				case 1: //A user was specified
					user, err := GitHubFetchUser(request[0])
					if err != nil {
						responseEmbed = NewErrorEmbed("GitHub Error", "There was an error finding info about that user.")
					} else {
						fields := []*discordgo.MessageEmbedField{}

						//Gather user info
						if user.Bio != nil {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Bio", Value: *user.Bio})
						}
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Username", Value: *user.Login})
						if user.Name != nil {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Name", Value: *user.Name})
						}
						if user.Company != nil {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Company", Value: *user.Company})
						}
						if *user.Blog != "" {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Blog", Value: *user.Blog})
						}
						if user.Location != nil {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Location", Value: *user.Location})
						}
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Public Repos", Value: strconv.Itoa(*user.PublicRepos)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Public Gists", Value: strconv.Itoa(*user.PublicGists)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Following", Value: strconv.Itoa(*user.Following)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Followers", Value: strconv.Itoa(*user.Followers)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "GitHub URL", Value: *user.HTMLURL})

						for i := 0; i < len(fields); i++ {
							debugLog(fields[i].Name+": "+fields[i].Value, false)
						}

						//Build embed about user
						responseEmbed = NewEmbed().
							SetTitle("GitHub User: " + *user.Login).
							SetImage(*user.AvatarURL).
							SetColor(0x24292D).MessageEmbed
						responseEmbed.Fields = fields
					}
				case 2: //A repo under a user was specified
					repo, err := GitHubFetchRepo(request[0], request[1])
					if err != nil {
						responseEmbed = NewErrorEmbed("GitHub Error", "There was an error finding info about that repo.")
					} else {
						fields := []*discordgo.MessageEmbedField{}

						//Gather repo info
						if repo.Description != nil && *repo.Description != "" {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Description", Value: *repo.Description})
						}
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Name", Value: *repo.FullName})
						if repo.Homepage != nil && *repo.Homepage != "" {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Homepage", Value: *repo.Homepage})
						}
						if len(repo.Topics) > 0 {
							fields = append(fields, &discordgo.MessageEmbedField{Name: "Topics", Value: strings.Join(repo.Topics, ", ")})
						}
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Default Branch", Value: *repo.DefaultBranch})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Is Fork", Value: strconv.FormatBool(*repo.Fork)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Forks", Value: strconv.Itoa(*repo.ForksCount)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Networks", Value: strconv.Itoa(*repo.NetworkCount)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Open Issues", Value: strconv.Itoa(*repo.OpenIssuesCount)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Stargazers", Value: strconv.Itoa(*repo.StargazersCount)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Subscribers", Value: strconv.Itoa(*repo.SubscribersCount)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Watchers", Value: strconv.Itoa(*repo.WatchersCount)})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "GitHub URL", Value: *repo.HTMLURL})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Clone URL", Value: *repo.CloneURL})
						fields = append(fields, &discordgo.MessageEmbedField{Name: "Git URL", Value: *repo.GitURL})

						for i := 0; i < len(fields); i++ {
							debugLog(fields[i].Name+": "+fields[i].Value, false)
						}

						//Build embed about repo
						responseEmbed = NewEmbed().
							SetTitle("GitHub Repo: " + *repo.FullName).
							SetColor(0x24292D).MessageEmbed
						responseEmbed.Fields = fields
					}
				default:
					responseEmbed = NewErrorEmbed("GitHub Error", "You got a little too specific there! Make sure to only specify either a user or a user/repo combination.")
				}
			} else {
				responseEmbed = NewErrorEmbed("GitHub Error", "You must specify a GitHub user or a GitHub repo to fetch info about.\n\nExamples:\n```"+botData.CommandPrefix+"github JoshuaDoes\n"+botData.CommandPrefix+"gh JoshuaDoes/clinet-discord```")
			}
		case "join":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
					responseEmbed = NewGenericEmbed("Clinet Voice", "Joined voice channel.")
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel to use before using the join command.")
			}
		case "leave":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					if voiceIsStreaming(guild.ID) {
						voiceStop(guild.ID)
					}
					err := voiceLeave(guild.ID, voiceState.ChannelID)
					if err != nil {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error leaving the voice channel.")
					} else {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Left voice channel.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the leave command.")
			}
		case "play":
			if updatedMessageEvent {
				//Todo: Remove this once I figure out how to detect if message update was user-triggered
				//Reason: If you use a YouTube/SoundCloud URL, Discord automatically updates the message with an embed
				//	As far as I know, bots have no way to know if this was a Discord- or user-triggered message update
				//I eventually want users to be able to edit their play command to change a now playing or a queue entry that was misspelled
				return
			}
			for guildData[guild.ID].VoiceData.IsPlaybackPreparing {
				//Wait for the handling of a previous playback command to finish
			}
			guildData[guild.ID].VoiceData.IsPlaybackPreparing = true
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
					break
				}
			}
			if foundVoiceChannel {
				if len(cmd) == 1 { //No query or URL was specified
					if voiceIsStreaming(guild.ID) {
						if len(message.Attachments) > 0 {
							for _, attachment := range message.Attachments {
								queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: message.Author}
								queueData.FillMetadata()
								guildData[guild.ID].QueueAdd(queueData)
							}
							responseEmbed = NewGenericEmbed("Clinet Voice", "Added the attached files to the guild queue.")
						} else {
							isPaused, _ := voiceGetPauseState(guild.ID)
							if isPaused {
								_, _ = voiceResume(guild.ID)
								responseEmbed = NewGenericEmbed("Clinet Voice", "Resumed the audio playback.")
							} else {
								responseEmbed = NewErrorEmbed("Clinet Voice Error", "The current audio is already playing.")
							}
						}
					} else {
						if len(message.Attachments) > 0 {
							for _, attachment := range message.Attachments {
								queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: message.Author}
								queueData.FillMetadata()
								guildData[guild.ID].QueueAdd(queueData)
							}
							responseEmbed = NewGenericEmbed("Clinet Voice", "Added the attached files to the guild queue. Use ``"+botData.CommandPrefix+"play`` to begin playback from the beginning of the queue.")
						} else {
							if guildData[guild.ID].AudioNowPlaying.MediaURL != "" {
								queueData := guildData[guild.ID].AudioNowPlaying
								queueData.FillMetadata()
								responseEmbed = queueData.GetQueueAddedEmbed()
								go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
							} else {
								if len(guildData[guild.ID].AudioQueue) > 0 {
									queueData := guildData[guild.ID].AudioQueue[0]
									queueData.FillMetadata()
									guildData[guild.ID].QueueRemove(0)
									responseEmbed = queueData.GetQueueAddedEmbed()
									go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
								} else {
									responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must specify either a YouTube search query or a YouTube/SoundCloud/direct URL to play.")
								}
							}
						}
					}
				} else if len(cmd) == 2 { //One-word query or URL was specified
					_, err := url.ParseRequestURI(cmd[1]) //Check to see if first parameter is URL
					if err != nil {                       //First parameter is not URL
						queryURL, err := voiceGetQuery(cmd[1])
						if err != nil {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error getting a result for the specified query.")
						} else {
							queueData := AudioQueueEntry{MediaURL: queryURL, Requester: message.Author, Type: "youtube"}
							queueData.FillMetadata()
							if voiceIsStreaming(guild.ID) {
								guildData[guild.ID].QueueAdd(queueData)
								responseEmbed = queueData.GetQueueAddedEmbed()
							} else {
								guildData[guild.ID].AudioNowPlaying = queueData
								responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
								go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
							}
						}
					} else { //First parameter is URL
						queueData := AudioQueueEntry{MediaURL: cmd[1], Requester: message.Author}
						queueData.FillMetadata()
						if voiceIsStreaming(guild.ID) {
							guildData[guild.ID].QueueAdd(queueData)
							responseEmbed = queueData.GetQueueAddedEmbed()
						} else {
							guildData[guild.ID].AudioNowPlaying = queueData
							responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
							go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
						}
					}
				} else if len(cmd) >= 3 { //Multi-word query was specified
					query := strings.Join(cmd[1:], " ") //Get the full search query without the play command
					queryURL, err := voiceGetQuery(query)
					if err != nil {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error getting a result for the specified query.")
					} else {
						queueData := AudioQueueEntry{MediaURL: queryURL, Requester: message.Author, Type: "youtube"}
						queueData.FillMetadata()
						if voiceIsStreaming(guild.ID) {
							guildData[guild.ID].QueueAdd(queueData)
							responseEmbed = queueData.GetQueueAddedEmbed()
						} else {
							guildData[guild.ID].AudioNowPlaying = queueData
							responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
							go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
						}
					}
				}
			} else {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel to use before using the play command.")
			}
			guildData[guild.ID].VoiceData.IsPlaybackPreparing = false
		case "stop":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					if voiceIsStreaming(guild.ID) {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Stopped the audio playback.")
						voiceStop(guild.ID)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					foundVoiceChannel = true
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the stop command.")
			}
		case "skip":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					if voiceIsStreaming(guild.ID) {
						voiceSkip(guild.ID)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					foundVoiceChannel = true
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the skip command.")
			}
		case "pause":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					isPaused, err := voicePause(guild.ID)
					if err != nil {
						if isPaused == false {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
						} else {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "The current audio is already paused.")
						}
					} else {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Paused the audio playback.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the pause command.")
			}
		case "resume":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					isPaused, err := voiceResume(guild.ID)
					if err != nil {
						if isPaused == false {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
						} else {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "The current audio is already playing.")
						}
					} else {
						responseEmbed = NewGenericEmbed("Clinet Voice", "Resumed the audio playback.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the resume command.")
			}
		case "repeat":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					if voiceIsStreaming(guild.ID) {
						switch guildData[guild.ID].VoiceData.RepeatLevel {
						case 0: //No Repeat
							guildData[guild.ID].VoiceData.RepeatLevel = 1
							responseEmbed = NewGenericEmbed("Clinet Voice", "The current guild queue will be repeated.")
						case 1: //Repeat Playlist
							guildData[guild.ID].VoiceData.RepeatLevel = 2
							responseEmbed = NewGenericEmbed("Clinet Voice", "The currently now playing audio will be repeated.")
						case 2: //Repeat Now Playing
							guildData[guild.ID].VoiceData.RepeatLevel = 0
							responseEmbed = NewGenericEmbed("Clinet Voice", "The current guild queue will play through as normal.")
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the repeat command.")
			}
		case "shuffle":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					foundVoiceChannel = true
					if voiceIsStreaming(guild.ID) {
						newAudioQueue := make([]AudioQueueEntry, len(guildData[guild.ID].AudioQueue))
						permutation := rand.Perm(len(guildData[guild.ID].AudioQueue))
						for i, v := range permutation {
							newAudioQueue[v] = guildData[guild.ID].AudioQueue[i]
						}
						guildData[guild.ID].AudioQueue = newAudioQueue

						responseEmbed = NewGenericEmbed("Clinet Voice", "The current guild queue has been shuffled.")
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the shuffle command.")
			}
		case "youtube", "yt":
			if len(cmd) > 1 {
				switch cmd[1] {
				case "search", "s":
					if guildData[guild.ID] == nil {
						guildData[guild.ID] = &GuildData{}
						guildData[guild.ID].VoiceData = VoiceData{}
					}
					for guildData[guild.ID].VoiceData.IsPlaybackPreparing {
						//Wait for the handling of a previous playback command to finish
					}
					foundVoiceChannel := false
					for _, voiceState := range guild.VoiceStates {
						if voiceState.UserID == message.Author.ID {
							foundVoiceChannel = true
							voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
							break
						}
					}
					if foundVoiceChannel {
						query := strings.Join(cmd[2:], " ") //Get the full search query without the search command
						if query == "" {
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "You must enter a search query to view the results of.")
						} else {
							_, guildFound := guildData[guild.ID]
							if !guildFound {
								guildData[guild.ID] = &GuildData{}
							}
							if guildData[guild.ID].YouTubeResults == nil {
								guildData[guild.ID].YouTubeResults = make(map[string]*YouTubeResultNav)
							}

							guildData[guild.ID].YouTubeResults[message.Author.ID] = &YouTubeResultNav{}
							page := guildData[guild.ID].YouTubeResults[message.Author.ID]
							err := page.Search(query)
							if err != nil {
								responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "There was an error finding search results for that query.")
							} else {
								commandList := "cli$yt select N - Selects result N"
								if page.PrevPageToken != "" {
									commandList += "\ncli$yt prev - Displays the results for the previous page"
								}
								if page.NextPageToken != "" {
									commandList += "\ncli$yt next - Displays the results for the next page"
								}
								commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

								results, _ := page.GetResults()
								responseEmbed = NewEmbed().
									SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
									SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
									SetColor(0xFF0000).MessageEmbed

								fields := []*discordgo.MessageEmbedField{}
								for i := 0; i < len(results); i++ {
									videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
									if err == nil {
										author := videoInfo.Author
										title := videoInfo.Title

										fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "\"" + title + "\" by " + author})
									}
								}
								fields = append(fields, commandListField)
								responseEmbed.Fields = fields
							}
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "You must join the voice channel to use before using the YouTube search command.")
					}
				case "next", "n", "+":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						page := guildData[guild.ID].YouTubeResults[message.Author.ID]
						err := page.Next()
						if err != nil {
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "There was an error finding the next page.")
						} else {
							commandList := "cli$yt select N - Selects result N"
							if page.PrevPageToken != "" {
								commandList += "\ncli$yt prev - Displays the results for the previous page"
							}
							if page.NextPageToken != "" {
								commandList += "\ncli$yt next - Displays the results for the next page"
							}
							commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

							results, _ := page.GetResults()
							responseEmbed = NewEmbed().
								SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
								SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
								SetColor(0xFF0000).MessageEmbed

							fields := []*discordgo.MessageEmbedField{}
							for i := 0; i < len(results); i++ {
								videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
								if err == nil {
									author := videoInfo.Author
									title := videoInfo.Title

									fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "\"" + title + "\" by " + author})
								}
							}
							fields = append(fields, commandListField)
							responseEmbed.Fields = fields
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				case "prev", "previous", "p", "-":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						page := guildData[guild.ID].YouTubeResults[message.Author.ID]
						err := page.Prev()
						if err != nil {
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "There was an error finding the previous page.")
						} else {
							commandList := "cli$yt select N - Selects result N"
							if page.PrevPageToken != "" {
								commandList += "\ncli$yt prev - Displays the results for the previous page"
							}
							if page.NextPageToken != "" {
								commandList += "\ncli$yt next - Displays the results for the next page"
							}
							commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

							results, _ := page.GetResults()
							responseEmbed = NewEmbed().
								SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
								SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
								SetColor(0xFF0000).MessageEmbed

							fields := []*discordgo.MessageEmbedField{}
							for i := 0; i < len(results); i++ {
								videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
								if err == nil {
									author := videoInfo.Author
									title := videoInfo.Title

									fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "\"" + title + "\" by " + author})
								}
							}
							fields = append(fields, commandListField)
							responseEmbed.Fields = fields
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				case "cancel", "c":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						guildData[guild.ID].YouTubeResults[message.Author.ID] = nil
						responseEmbed = NewGenericEmbed("Clinet YouTube Search", "Cancelled the search.")
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				case "select":
					if guildData[guild.ID].YouTubeResults[message.Author.ID] != nil {
						page := guildData[guild.ID].YouTubeResults[message.Author.ID]
						results, _ := page.GetResults()

						selection, err := strconv.Atoi(cmd[2])
						if err != nil { //Specified selection is not a valid integer
							responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "``"+cmd[2]+"`` is not a valid number.")
						} else {
							if selection > len(results) || selection <= 0 {
								responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "An invalid selection was specified.")
							} else {
								result := results[selection-1]
								resultURL := "https://youtube.com/watch?v=" + result.Id.VideoId

								queueData := AudioQueueEntry{MediaURL: resultURL, Requester: message.Author, Type: "youtube"}
								queueData.FillMetadata()
								if voiceIsStreaming(guild.ID) {
									guildData[guild.ID].QueueAdd(queueData)
									responseEmbed = queueData.GetQueueAddedEmbed()
								} else {
									guildData[guild.ID].AudioNowPlaying = queueData
									responseEmbed = queueData.GetNowPlayingEmbed()
									go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
								}
							}
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet YouTube Search Error", "No search is in progress.")
					}
				}
			}
		case "queue":
			if len(cmd) > 1 {
				switch cmd[1] {
				case "help":
					responseEmbed = NewEmbed().
						SetTitle(botData.BotName+" - Queue Help").
						SetDescription("A list of available queue commands for "+botData.BotName+".").
						AddField(botData.CommandPrefix+"queue help", "Displays this help message.").
						AddField(botData.CommandPrefix+"queue clear", "Clears the current queue.").
						AddField(botData.CommandPrefix+"queue list", "Lists all entries in the queue.").
						AddField(botData.CommandPrefix+"queue remove (entry)", "Removes a specified entry from the queue.").
						SetColor(0xFAFAFA).MessageEmbed
				case "clear":
					if guildData[guild.ID].AudioQueue == nil {
						guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
					}
					if len(guildData[guild.ID].AudioQueue) > 0 {
						guildData[guild.ID].QueueClear()
						responseEmbed = NewGenericEmbed("Clinet Queue", "Cleared the queue.")
					} else {
						responseEmbed = NewErrorEmbed("Clinet Queue Error", "There are no entries in the queue to clear.")
					}
				case "list":
					if guildData[guild.ID].AudioQueue == nil {
						guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
					}
					if len(guildData[guild.ID].AudioQueue) > 0 {
						queueList := ""
						for queueEntryNumber, queueEntry := range guildData[guild.ID].AudioQueue {
							displayNumber := strconv.Itoa(queueEntryNumber + 1)
							if queueList != "" {
								queueList += "\n"
							}
							switch queueEntry.Type {
							case "youtube", "soundcloud":
								queueList += displayNumber + ". ``" + queueEntry.Title + "`` by ``" + queueEntry.Author + "``\n\tRequested by " + queueEntry.Requester.String()
							default:
								queueList += displayNumber + ". ``" + queueEntry.MediaURL + "``\n\tRequested by " + queueEntry.Requester.String()
							}
						}
						responseEmbed = NewGenericEmbed("Queue for "+guild.Name, queueList)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Queue Error", "There are no entries in the queue to list.")
					}
				case "remove":
					if guildData[guild.ID].AudioQueue == nil {
						guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
					}
					if len(cmd) > 2 {
						invalidQueueEntry := ""
						for _, queueEntry := range cmd[2:] { //Range over all specified queue entries
							queueEntryNumber, err := strconv.Atoi(queueEntry)
							if err != nil { //Specified queue entry is not a valid integer
								invalidQueueEntry = queueEntry
								break
							} else {
								queueEntryNumber -= 1 //Compensate for 0-index
							}

							if queueEntryNumber > len(guildData[guild.ID].AudioQueue) || queueEntryNumber < 0 {
								invalidQueueEntry = queueEntry
								break
							}
						}
						if invalidQueueEntry != "" {
							responseEmbed = NewErrorEmbed("Clinet Queue Error", invalidQueueEntry+" is not a valid queue entry.")
						} else {
							var newAudioQueue []AudioQueueEntry
							for queueEntryN, queueEntry := range guildData[guild.ID].AudioQueue {
								keepQueueEntry := true
								for _, removedQueueEntry := range cmd[2:] {
									removedQueueEntryNumber, _ := strconv.Atoi(removedQueueEntry)
									removedQueueEntryNumber -= 1
									if queueEntryN == removedQueueEntryNumber {
										keepQueueEntry = false
										break
									}
								}
								if keepQueueEntry {
									newAudioQueue = append(newAudioQueue, queueEntry)
								}
							}

							guildData[guild.ID].AudioQueue = newAudioQueue

							if len(cmd) > 3 {
								responseEmbed = NewGenericEmbed("Clinet Queue", "Successfully removed the specified queue entries.")
							} else {
								responseEmbed = NewGenericEmbed("Clinet Queue", "Successfully removed the specified queue entry.")
							}
						}
					} else {
						responseEmbed = NewErrorEmbed("Clinet Queue Error", "You must specify which entries to remove from the queue.")
					}
				}
			} else {
				if guildData[guild.ID].AudioQueue == nil {
					guildData[guild.ID].AudioQueue = make([]AudioQueueEntry, 0)
				}
				if len(guildData[guild.ID].AudioQueue) > 0 {
					queueList := ""
					for queueEntryNumber, queueEntry := range guildData[guild.ID].AudioQueue {
						displayNumber := strconv.Itoa(queueEntryNumber + 1)
						if queueList != "" {
							queueList += "\n"
						}
						switch queueEntry.Type {
						case "youtube", "soundcloud":
							queueList += displayNumber + ". ``" + queueEntry.Title + "`` by ``" + queueEntry.Author + "``\n\tRequested by " + queueEntry.Requester.String()
						default:
							queueList += displayNumber + ". ``" + queueEntry.MediaURL + "``\n\tRequested by " + queueEntry.Requester.String()
						}
					}
					responseEmbed = NewGenericEmbed("Queue for "+guild.Name, queueList)
				} else {
					responseEmbed = NewErrorEmbed("Clinet Queue Error", "There are no entries in the queue to list.")
				}
			}
		case "nowplaying", "np":
			foundVoiceChannel := false
			for _, voiceState := range guild.VoiceStates {
				if voiceState.UserID == message.Author.ID {
					if voiceIsStreaming(guild.ID) {
						//Create and display now playing embed
						responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingDurationEmbed(guildData[guild.ID].VoiceData.StreamingSession)
					} else {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There is no audio currently playing.")
					}
					foundVoiceChannel = true
					break
				}
			}
			if foundVoiceChannel == false {
				responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the now playing command.")
			}
		}
	} else {
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
			if len(botData.CustomResponses) > 0 {
				for _, response := range botData.CustomResponses {
					regexpMatched, _ := regexp.MatchString(response.Expression, query)
					if regexpMatched {
						random := rand.Intn(len(response.Responses))
						responseEmbed = NewGenericEmbed("Clinet Response", response.Responses[random].Response)
						usedCustomResponse = true
					}
				}
			}
			if usedCustomResponse == false {

				//Experimental - Use regex for natural language-based commands
				regexCmdPlayComp, err := regexp.Compile(regexCmdPlay)
				if err != nil {
					panic(err)
				}
				match := regexCmdPlayComp.FindAllString(query, 1) //Get a slice of the results
				if len(match) > 0 {
					queryURL, err := voiceGetQuery(match[0])
					if err != nil {
						responseEmbed = NewErrorEmbed("Clinet Voice Error", "There was an error getting a result for the specified query.")
					} else {
						foundVoiceChannel := false
						for _, voiceState := range guild.VoiceStates {
							if voiceState.UserID == message.Author.ID {
								foundVoiceChannel = true
								voiceJoin(session, guild.ID, voiceState.ChannelID, message.ChannelID)
								break
							}
						}
						if foundVoiceChannel {
							queueData := AudioQueueEntry{MediaURL: queryURL, Requester: message.Author, Type: "youtube"}
							queueData.FillMetadata()
							if voiceIsStreaming(guild.ID) {
								guildData[guild.ID].QueueAdd(queueData)
								responseEmbed = queueData.GetQueueAddedEmbed()
							} else {
								guildData[guild.ID].AudioNowPlaying = queueData
								responseEmbed = guildData[guild.ID].AudioNowPlaying.GetNowPlayingEmbed()
								go voicePlayWrapper(session, guild.ID, message.ChannelID, queueData.MediaURL)
							}
						} else {
							responseEmbed = NewErrorEmbed("Clinet Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the play command.")
						}
					}
				} else { //End experimental

					responseEmbed, err = queryDuckDuckGo(query)
					if err != nil {
						responseEmbed, err = queryWolframAlpha(query)
						if err != nil {
							responseEmbed = NewErrorEmbed("Query Error", "There was an error finding the data you requested.")
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
		fixedEmbed.InlineAllFields().Truncate()
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
