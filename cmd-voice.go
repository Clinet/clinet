package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rylio/ytdl"
)

func commandVoiceJoin(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
			return NewGenericEmbed("Voice", "Joined the voice channel.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel to use before using the join command.")
}

func commandVoiceLeave(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if guildData[env.Guild.ID].VoiceData.VoiceConnection == nil {
		return NewErrorEmbed("Voice Error", botData.BotName+" is not currently in a voice channel.")
	}

	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID && voiceState.ChannelID == guildData[env.Guild.ID].VoiceData.VoiceConnection.ChannelID {
			voiceStop(env.Guild.ID)
			err := voiceLeave(env.Guild.ID, voiceState.ChannelID)
			if err != nil {
				return NewErrorEmbed("Voice Error", "There was an error leaving the voice channel.")
			}
			return NewGenericEmbed("Voice", "Left the voice channel.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the leave command.")
}

func commandPlay(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if env.UpdatedMessageEvent {
		//Todo: Remove this once I figure out how to detect if message update was user-triggered
		//Reason: If you use a YouTube/SoundCloud URL, Discord automatically updates the message with an embed
		//As far as I know, bots have no way to know if this was a Discord- or user-triggered message update
		//I eventually want users to be able to edit their play command to change a now playing or a queue entry that was misspelled
		return nil
	}

	for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
		//Wait for the handling of a previous playback command to finish
	}
	foundVoiceChannel := false
	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID {
			if guildData[env.Guild.ID].VoiceData.VoiceConnection != nil && voiceState.ChannelID != guildData[env.Guild.ID].VoiceData.VoiceConnection.ChannelID {
				return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the play command.")
			}
			foundVoiceChannel = true
			voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
			break
		}
	}
	if !foundVoiceChannel {
		return NewErrorEmbed("Voice Error", "You must join the voice channel to use before using the play command.")
	}
	//Prevent other play commands in this voice session from messing up this process
	guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = true

	if len(args) >= 1 { //Query or URL was specified
		_, err := url.ParseRequestURI(args[0]) //Check to see if the first parameter is a URL
		if err != nil {                        //First parameter is not a URL
			queryURL, err := voiceGetQuery(strings.Join(args, " "))
			if err != nil {
				guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
				return NewErrorEmbed("Voice Error", "There was an error getting a result for the specified query.")
			}
			queueData := AudioQueueEntry{MediaURL: queryURL, Requester: env.Message.Author, Type: "youtube"}
			errEmbed := queueData.FillMetadata()
			if errEmbed != nil {
				guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false
				return errEmbed
			}
			if voiceIsStreaming(env.Guild.ID) {
				guildData[env.Guild.ID].QueueAdd(queueData)
				guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
				return queueData.GetQueueAddedEmbed()
			}
			guildData[env.Guild.ID].AudioNowPlaying = queueData
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)
			return queueData.GetNowPlayingEmbed()
		}

		//First parameter is a URL
		queueData := AudioQueueEntry{MediaURL: args[0], Requester: env.Message.Author}
		errEmbed := queueData.FillMetadata()
		if errEmbed != nil {
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false
			return errEmbed
		}
		if voiceIsStreaming(env.Guild.ID) {
			guildData[env.Guild.ID].QueueAdd(queueData)
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			return queueData.GetQueueAddedEmbed()
		}
		guildData[env.Guild.ID].AudioNowPlaying = queueData
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)
		return queueData.GetNowPlayingEmbed()
	}

	if voiceIsStreaming(env.Guild.ID) {
		if len(env.Message.Attachments) > 0 {
			for _, attachment := range env.Message.Attachments {
				queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: env.Message.Author}
				errEmbed := queueData.FillMetadata()
				if errEmbed != nil {
					guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false
					return errEmbed
				}
				guildData[env.Guild.ID].QueueAdd(queueData)
			}
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			return NewGenericEmbed("Voice", "Added the attached files to the guild queue.")
		}
		isPaused, _ := voiceGetPauseState(env.Guild.ID)
		if isPaused {
			_, _ = voiceResume(env.Guild.ID)
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
			return NewGenericEmbed("Voice", "Resumed the audio playback.")
		}
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		return NewErrorEmbed("Voice Error", "Already playing audio.")
	}
	if len(env.Message.Attachments) > 0 {
		for _, attachment := range env.Message.Attachments {
			queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: env.Message.Author}
			errEmbed := queueData.FillMetadata()
			if errEmbed != nil {
				guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
				return errEmbed
			}
			guildData[env.Guild.ID].QueueAdd(queueData)
		}
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		return NewGenericEmbed("Voice", "Added the attached files to the guild queue. Use ``"+env.BotPrefix+"play`` to begin playback from the beginning of the queue.")
	}
	if guildData[env.Guild.ID].AudioNowPlaying.MediaURL != "" {
		queueData := guildData[env.Guild.ID].AudioNowPlaying
		errEmbed := queueData.FillMetadata()
		if errEmbed != nil {
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false
			return errEmbed
		}
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Message.ChannelID, queueData.MediaURL)
		return queueData.GetQueueAddedEmbed()
	}
	if len(guildData[env.Guild.ID].AudioQueue) > 0 {
		queueData := guildData[env.Guild.ID].AudioQueue[0]
		errEmbed := queueData.FillMetadata()
		if errEmbed != nil {
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false
			return errEmbed
		}
		guildData[env.Guild.ID].QueueRemove(0)
		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Message.ChannelID, queueData.MediaURL)
		return queueData.GetQueueAddedEmbed()
	}

	guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run
	return NewErrorEmbed("Voice Error", "Could not find any audio to play.")
}

func commandStop(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if guildData[env.Guild.ID].VoiceData.VoiceConnection == nil {
		return NewErrorEmbed("Voice Error", botData.BotName+" is not currently in a voice channel.")
	}

	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID && voiceState.ChannelID == guildData[env.Guild.ID].VoiceData.VoiceConnection.ChannelID {
			if voiceIsStreaming(env.Guild.ID) {
				voiceStop(env.Guild.ID)
				return NewGenericEmbed("Voice", "Stopped the audio playback.")
			}
			return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the "+env.Command+" command.")
}

func commandSkip(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if guildData[env.Guild.ID].VoiceData.VoiceConnection == nil {
		return NewErrorEmbed("Voice Error", botData.BotName+" is not currently in a voice channel.")
	}

	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID && voiceState.ChannelID == guildData[env.Guild.ID].VoiceData.VoiceConnection.ChannelID {
			if voiceIsStreaming(env.Guild.ID) {
				voiceSkip(env.Guild.ID)
				return nil
			}
			return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the "+env.Command+" command.")
}

func commandPause(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if guildData[env.Guild.ID].VoiceData.VoiceConnection == nil {
		return NewErrorEmbed("Voice Error", botData.BotName+" is not currently in a voice channel.")
	}

	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID && voiceState.ChannelID == guildData[env.Guild.ID].VoiceData.VoiceConnection.ChannelID {
			isPaused, err := voicePause(env.Guild.ID)
			if err != nil {
				if isPaused {
					return NewErrorEmbed("Voice Error", "Already playing audio.")
				}
				return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
			}
			return NewGenericEmbed("Voice", "Paused the audio playback.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" is in before using the "+env.Command+" command.")
}

func commandResume(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if guildData[env.Guild.ID].VoiceData.VoiceConnection == nil {
		return NewErrorEmbed("Voice Error", botData.BotName+" is not currently in a voice channel.")
	}

	for _, voiceState := range env.Guild.VoiceStates {
		if voiceState.UserID == env.Message.Author.ID && voiceState.ChannelID == guildData[env.Guild.ID].VoiceData.VoiceConnection.ChannelID {
			isPaused, err := voiceResume(env.Guild.ID)
			if err != nil {
				if isPaused {
					return NewErrorEmbed("Voice Error", "Already playing audio.")
				}
				return NewErrorEmbed("Voice Error", "There is no audio currently playing.")
			}
			return NewGenericEmbed("Voice", "Resumed the audio playback.")
		}
	}
	return NewErrorEmbed("Voice Error", "You must join the voice channel "+botData.BotName+" to use before using the "+env.Command+" command.")
}

func commandVolume(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//Disabled until further notice, real-time volume control using hrabin/opus and manually adjusting samples results in static noise distortion with the correct volume
	/*
		volume, err := strconv.Atoi(args[0])
		if err != nil {
			return NewErrorEmbed("Volume Error", "``"+args[0]+"`` is not a valid number.")
		}

		if volume < 0 || volume > 100 {
			return NewErrorEmbed("Volume Error", "You must specify a volume level from 0 to 100, with 100 being normal volume.")
		}

		if guildData[env.Guild.ID].VoiceData.EncodingOptions == nil {
			guildData[env.Guild.ID].VoiceData.EncodingOptions = encodeOptionsPresetHigh
		}
		guildData[env.Guild.ID].VoiceData.EncodingOptions.Volume = float64(volume) * 0.01
		return NewErrorEmbed("Volume", "Set the volume for audio playback to "+args[0]+".")
	*/

	return NewGenericEmbed("Volume", "Volume adjustment in real time via this command is disabled at this time. While attempts proved to successfully change the volume, it was accompanied by static noise distortion and thus is not ready for production.\n"+
		"If you wish to change your perceived volume of Clinet, consider using Discord's per-user volume control (right click Clinet on desktop/web or tap on Clinet in the user list on mobile to find it). Not only does it do what you want, but it doesn't have to ruin everyone else's high quality audio experience!\n"+
		"If you would like to help with attempts to change the volume in real time, make sure to join the [Clinet Discord server](https://discord.gg/qkbKEWT).")
}

func commandRepeat(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch guildData[env.Guild.ID].VoiceData.RepeatLevel {
	case 0: //No repeat
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 1
		return NewGenericEmbed("Voice", "The queue will be repeated on a loop.")
	case 1: //Repeat the current queue
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 2
		return NewGenericEmbed("Voice", "The now playing entry will be repeated on a loop.")
	case 2: //Repeat what's in the now playing slot
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 0
		return NewGenericEmbed("Voice", "The queue will play through as normal.")
	}
	return nil
}

func commandShuffle(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	guildData[env.Guild.ID].VoiceData.Shuffle = !guildData[env.Guild.ID].VoiceData.Shuffle
	if guildData[env.Guild.ID].VoiceData.Shuffle {
		return NewGenericEmbed("Voice", "The queue will be shuffled around in a random order while playing.")
	}
	return NewGenericEmbed("Voice", "The queue will play through as normal.")
}

func commandYouTube(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	page := &YouTubeResultNav{}

	switch args[0] {
	case "search", "s":
		query := strings.Join(args[1:], " ")
		if query == "" {
			return NewErrorEmbed("YouTube Error", "You must enter a search query to use before using the "+args[0]+" command.")
		}

		if guildData[env.Guild.ID].YouTubeResults == nil {
			guildData[env.Guild.ID].YouTubeResults = make(map[string]*YouTubeResultNav)
		}

		guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID] = &YouTubeResultNav{}

		page = guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		err := page.Search(query)
		if err != nil {
			return NewErrorEmbed("YouTube Error", "There was an error getting a result for the specified query.")
		}
	case "next", "n", "forward", "+":
		if guildData[env.Guild.ID].YouTubeResults == nil {
			return NewErrorEmbed("YouTube Error", "No search session is in progress.")
		}

		page = guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		err := page.Next()
		if err != nil {
			return NewErrorEmbed("YouTube Error", "There was an error finding the next page.")
		}
	case "prev", "previous", "p", "back", "-":
		if guildData[env.Guild.ID].YouTubeResults == nil {
			return NewErrorEmbed("YouTube Error", "No search session is in progress.")
		}

		page = guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		err := page.Prev()
		if err != nil {
			return NewErrorEmbed("YouTube Error", "There was an error finding the previous page.")
		}
	case "cancel", "c":
		if guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID] != nil {
			guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID] = nil
			return NewGenericEmbed("YouTube", "Cancelled the search session.")
		}
		return NewErrorEmbed("YouTube Error", "No search session is in progress.")
	case "select", "choose", "play":
		if guildData[env.Guild.ID].YouTubeResults == nil {
			return NewErrorEmbed("YouTube Error", "No search session is in progress.")
		}
		if len(args) < 2 {
			return NewErrorEmbed("YouTube Error", "You must specify which search result to select.")
		}

		page = guildData[env.Guild.ID].YouTubeResults[env.Message.Author.ID]
		results, _ := page.GetResults()

		selection, err := strconv.Atoi(args[1])
		if err != nil {
			return NewErrorEmbed("YouTube Error", "``"+args[1]+"`` is not a valid number.")
		}
		if selection > len(results) || selection <= 0 {
			return NewErrorEmbed("YouTube Error", "An invalid selection was specified.")
		}

		for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
			//Wait for the handling of a previous playback command to finish
		}
		foundVoiceChannel := false
		for _, voiceState := range env.Guild.VoiceStates {
			if voiceState.UserID == env.Message.Author.ID {
				foundVoiceChannel = true
				voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
				break
			}
		}
		if !foundVoiceChannel {
			return NewErrorEmbed("YouTube Error", "You must join the voice channel to use before using the "+args[0]+" command.")
		}

		result := results[selection-1]
		resultURL := "https://youtube.com/watch?v=" + result.Id.VideoId

		queueData := AudioQueueEntry{MediaURL: resultURL, Requester: env.Message.Author, Type: "youtube"}
		errEmbed := queueData.FillMetadata()
		if errEmbed != nil {
			guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false
			return errEmbed
		}
		if voiceIsStreaming(env.Guild.ID) {
			guildData[env.Guild.ID].QueueAdd(queueData)
			return queueData.GetQueueAddedEmbed()
		}
		guildData[env.Guild.ID].AudioNowPlaying = queueData
		go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)
		return queueData.GetNowPlayingEmbed()
	default:
		return NewErrorEmbed("YouTube Error", "Unknown command ``"+args[0]+"``.")
	}

	commandList := env.BotPrefix + env.Command + " play N - Plays result N"
	if page.PrevPageToken != "" {
		commandList += "\n" + env.BotPrefix + env.Command + " prev - Displays the results for the previous page"
	}
	if page.NextPageToken != "" {
		commandList += "\n" + env.BotPrefix + env.Command + " next - Displays the results for the next page"
	}
	commandList += "\n" + env.BotPrefix + env.Command + " cancel - Cancels the search session"
	commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

	results, _ := page.GetResults()
	responseEmbed := NewEmbed().
		SetTitle("YouTube Search Results - Page " + strconv.Itoa(page.PageNumber)).
		SetDescription(strconv.FormatInt(page.TotalResults, 10) + " results for \"" + page.Query + "\"").
		SetColor(0xFF0000).MessageEmbed

	fields := []*discordgo.MessageEmbedField{}
	for i := 0; i < len(results); i++ {
		videoInfo, err := ytdl.GetVideoInfo("https://youtube.com/watch?v=" + results[i].Id.VideoId)
		if err != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "Error fetching info for [this video](https://youtube.com/watch?v=" + results[i].Id.VideoId + ")"})
		} else {
			author := videoInfo.Author
			title := videoInfo.Title

			fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "[" + title + "](https://youtube.com/watch?v=" + results[i].Id.VideoId + ") by **" + author + "**"})
		}
	}
	fields = append(fields, commandListField)
	responseEmbed.Fields = fields

	return responseEmbed
}

func commandQueue(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(args) >= 1 {
		switch args[0] {
		case "clear":
			if len(guildData[env.Guild.ID].AudioQueue) > 0 {
				guildData[env.Guild.ID].QueueClear()
				return NewGenericEmbed("Queue", "Cleared the queue.")
			}
			return NewErrorEmbed("Queue Error", "There are no entries in the queue to clear.")
		case "remove":
			if len(args) == 1 {
				return NewErrorEmbed("Queue Error", "You must specify which queue entries to remove.")
			}

			for _, queueEntry := range args[1:] {
				queueEntryNumber, err := strconv.Atoi(queueEntry)
				if err != nil {
					return NewErrorEmbed("Queue Error", "``"+queueEntry+"`` is not a valid number.")
				}
				queueEntryNumber--

				if queueEntryNumber > len(guildData[env.Guild.ID].AudioQueue) || queueEntryNumber < 0 {
					return NewErrorEmbed("Queue Error", "``"+queueEntry+"`` is not a valid queue entry.")
				}
			}

			var newAudioQueue []AudioQueueEntry
			for queueEntryN, queueEntry := range guildData[env.Guild.ID].AudioQueue {
				keepQueueEntry := true
				for _, removedQueueEntry := range args[1:] {
					removedQueueEntryNumber, _ := strconv.Atoi(removedQueueEntry)
					removedQueueEntryNumber--
					if queueEntryN == removedQueueEntryNumber {
						keepQueueEntry = false
						break
					}
				}
				if keepQueueEntry {
					newAudioQueue = append(newAudioQueue, queueEntry)
				}
			}

			guildData[env.Guild.ID].AudioQueue = newAudioQueue

			if len(args) > 2 {
				return NewGenericEmbed("Queue", "Successfully removed the specified queue entries.")
			}
			return NewGenericEmbed("Queue", "Successfully removed the specified queue entry.")
		case "copy":
			if len(args) == 1 {
				return NewErrorEmbed("Queue Error", "You must specify which guild queue(s) to copy.")
			}

			for _, guildID := range args[1:] {
				if _, exists := guildData[guildID]; exists == false {
					return NewErrorEmbed("Queue Error", "The guild ID ``"+guildID+"`` does not point to a known guild.")
				}
			}

			copiedGuilds := make([]string, 0)
			for _, guildID := range args[1:] {
				if guild, exists := guildData[guildID]; exists { //Just in case it doesn't exist anymore when we reach this point, we all know how edge cases go
					if guild.AudioNowPlaying.MediaURL != "" || len(guild.AudioQueue) > 0 {
						if guild.AudioNowPlaying.MediaURL != "" {
							guildData[env.Guild.ID].AudioQueue = append(guildData[env.Guild.ID].AudioQueue, guild.AudioNowPlaying)
						}
						if len(guild.AudioQueue) > 0 {
							for i := 0; i < len(guild.AudioQueue); i++ {
								guildData[env.Guild.ID].AudioQueue = append(guildData[env.Guild.ID].AudioQueue, guild.AudioQueue[i])
							}
						}

						guildState, _ := botData.DiscordSession.State.Guild(guildID)
						copiedGuilds = append(copiedGuilds, "**"+guildState.Name+"**")
					}
				}
			}

			if len(copiedGuilds) == 1 {
				return NewGenericEmbed("Queue", "Successfully copied the queue from "+copiedGuilds[0]+".")
			}
			return NewGenericEmbed("Queue", "Successfully copied the queue from the following guilds:\n"+strings.Join(copiedGuilds, "\n"))
		}
	}

	queueList := "There are no entries in the queue."
	for queueEntryNumber, queueEntry := range guildData[env.Guild.ID].AudioQueue {
		displayNumber := strconv.Itoa(queueEntryNumber + 1)
		if queueList != "There are no entries in the queue." {
			queueList += "\n"
		} else {
			queueList = ""
		}
		queueList += displayNumber + ". ``" + secondsToHuman(queueEntry.Duration) + "`` "
		if queueEntry.Title != "" && queueEntry.Author != "" {
			queueList += "[" + queueEntry.Title + "](" + queueEntry.MediaURL + ") by **" + queueEntry.Author + "**"
		} else {
			queueList += queueEntry.MediaURL
		}
		queueList += "\nRequested by <@!" + queueEntry.Requester.ID + ">"
	}

	queueEmbed := NewEmbed().
		SetTitle("Queue for " + env.Guild.Name).
		SetDescription(queueList).
		SetColor(0x1C1C1C)

	nowPlaying := guildData[env.Guild.ID].AudioNowPlaying
	nowPlayingField := &discordgo.MessageEmbedField{
		Name:  "Now Playing",
		Value: "There is no audio currently playing.",
	}

	if voiceIsStreaming(env.Guild.ID) {
		switch nowPlaying.Type {
		case "youtube":
			nowPlayingField.Name += " from YouTube"
			nowPlayingField.Value = fmt.Sprintf("[%s](%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
			queueEmbed.SetThumbnail(nowPlaying.ThumbnailURL)
		case "soundcloud":
			nowPlayingField.Name += " from SoundCloud"
			nowPlayingField.Value = fmt.Sprintf("[%s](%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
			queueEmbed.SetThumbnail(nowPlaying.ThumbnailURL)
		default:
			if nowPlaying.Author != "" && nowPlaying.Title != "" {
				nowPlayingField.Value = fmt.Sprintf("[%s](%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
			} else {
				nowPlayingField.Value = fmt.Sprintf("%s\nRequested by <@!%s>", nowPlaying.MediaURL, nowPlaying.Requester.ID)
			}
		}

	}
	queueEmbed.Fields = append(queueEmbed.Fields, nowPlayingField)

	return queueEmbed.MessageEmbed
}

func commandNowPlaying(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if voiceIsStreaming(env.Guild.ID) {
		return guildData[env.Guild.ID].AudioNowPlaying.GetNowPlayingDurationEmbed(guildData[env.Guild.ID].VoiceData.StreamingSession)
	}
	return NewErrorEmbed("Now Playing Error", "There is no audio currently playing.")
}
