package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/go-playground/colors"
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

	//Update channel ID to send voice messages to
	guildData[env.Guild.ID].VoiceData.ChannelIDJoinedFrom = env.Channel.ID

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

	if len(env.Message.Attachments) > 0 {
		waitEmbed := NewEmbed().
			SetTitle("Voice").
			SetDescription("Please wait a moment as we add all " + strconv.Itoa(len(env.Message.Attachments)) + " attachments to the queue...\n\nThe first result added will automatically begin playing. During this process, it may feel as if other commands are slow or don't work; give them some time to process.").
			SetColor(0x1DB954).MessageEmbed
		botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, waitEmbed)

		for i, attachment := range env.Message.Attachments {
			//Give a chance for other commands waiting in line to execute
			guildData[env.Guild.ID].Unlock()
			guildData[env.Guild.ID].Lock()

			queueData := AudioQueueEntry{MediaURL: attachment.URL, Requester: env.Message.Author}
			errEmbed := queueData.FillMetadata()
			if errEmbed != nil {
				botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, NewErrorEmbed("Voice Error", "Error finding track info for attachment "+strconv.Itoa(i+1)+"."))
			}

			if voiceIsStreaming(env.Guild.ID) {
				guildData[env.Guild.ID].QueueAdd(queueData)
				continue
			}

			guildData[env.Guild.ID].AudioNowPlaying = queueData
			go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)

			botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, queueData.GetNowPlayingEmbed())
		}

		guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing = false //We're done so we should allow the next play command to run

		return NewGenericEmbedAdvanced("Voice", "Finished adding all "+strconv.Itoa(len(env.Message.Attachments))+" attachments to the queue.", 0x1DB954)
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
		return queueData.GetNowPlayingEmbed()
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
		return queueData.GetNowPlayingEmbed()
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
	if len(args) > 0 {
		switch strings.Join(args, " ") {
		case "normal", "norm", "disable", "d", "0", "zero":
			guildData[env.Guild.ID].VoiceData.RepeatLevel = 0
			return NewGenericEmbed("Voice", "The queue will now play through as normal.")
		case "queue", "list", "queue list", "q", "l", "1", "one":
			guildData[env.Guild.ID].VoiceData.RepeatLevel = 1
			return NewGenericEmbed("Voice", "The queue will now be repeated on a loop.")
		case "nowplaying", "now playing", "now", "playing", "np", "n", "enable", "e", "2", "two":
			guildData[env.Guild.ID].VoiceData.RepeatLevel = 2
			return NewGenericEmbed("Voice", "The now playing entry will now be repeated on a loop.")
		}
	}
	switch guildData[env.Guild.ID].VoiceData.RepeatLevel {
	case 0: //No repeat
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 1
		return NewGenericEmbed("Voice", "The queue will now be repeated on a loop.")
	case 1: //Repeat the current queue
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 2
		return NewGenericEmbed("Voice", "The now playing entry will now be repeated on a loop.")
	case 2: //Repeat what's in the now playing slot
		guildData[env.Guild.ID].VoiceData.RepeatLevel = 0
		return NewGenericEmbed("Voice", "The queue will now play through as normal.")
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
			return NewGenericEmbedAdvanced("YouTube", "Cancelled the search session.", 0xFF0000)
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

		//Update channel ID to send voice messages to
		guildData[env.Guild.ID].VoiceData.ChannelIDJoinedFrom = env.Channel.ID

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

	results, err := page.GetResults()
	if err != nil {
		return NewErrorEmbed("YouTube Error", "No search results were found.")
	}
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

func commandSpotify(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	page := &SpotifyResultNav{}

	switch args[0] {
	case "search", "s":
		query := strings.Join(args[1:], " ")
		if query == "" {
			return NewErrorEmbed("Spotify Error", "You must enter a search query to use before using the "+args[0]+" command.")
		}

		if guildData[env.Guild.ID].SpotifyResults == nil {
			guildData[env.Guild.ID].SpotifyResults = make(map[string]*SpotifyResultNav)
		}

		guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID] = &SpotifyResultNav{}

		page = guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID]
		err := page.Search(query)
		if err != nil {
			return NewErrorEmbed("Spotify Error", "There was an error getting a result for the specified query.")
		}
	case "playlist", "list":
		playlistURL := strings.Join(args[1:], " ")
		if playlistURL == "" {
			return NewErrorEmbed("Spotify Error", "You must enter a playlist URL to use before using the "+args[0]+" command.")
		}

		if guildData[env.Guild.ID].SpotifyResults == nil {
			guildData[env.Guild.ID].SpotifyResults = make(map[string]*SpotifyResultNav)
		}

		guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID] = &SpotifyResultNav{}
		guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID].GuildID = env.Guild.ID

		waitEmbed := NewEmbed().
			SetTitle("Spotify").
			SetDescription("Please wait a while as we fetch the tracks from the specified playlist...\n\nDuring this process, it may feel as if other commands are slow or don't work; give them some time to process.\nYou may cancel at any moment with ``" + env.BotPrefix + env.Command + " cancel``. Once cancelled, the tracks gathered so far will still be displayed.").
			SetColor(0x1DB954).MessageEmbed
		botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, waitEmbed)

		page = guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID]
		err := page.Playlist(playlistURL)
		if err != nil {
			return NewErrorEmbed("Spotify Error", "There was an error getting a result for the specified playlist.")
		}
	case "next", "n", "forward", "+":
		if guildData[env.Guild.ID].SpotifyResults == nil {
			return NewErrorEmbed("Spotify Error", "No search session is in progress.")
		}

		page = guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID]
		err := page.Next()
		if err != nil {
			return NewErrorEmbed("Spotify Error", "There was an error finding the next page.")
		}
	case "prev", "previous", "p", "back", "-":
		if guildData[env.Guild.ID].SpotifyResults == nil {
			return NewErrorEmbed("Spotify Error", "No search session is in progress.")
		}

		page = guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID]
		err := page.Prev()
		if err != nil {
			return NewErrorEmbed("Spotify Error", "There was an error finding the previous page.")
		}
	case "cancel", "c":
		page = guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID]
		if page == nil {
			return NewErrorEmbed("Spotify Error", "No Spotify session is in progress.")
		}

		if page.AddingAll {
			page.Cancelled = true
			return NewGenericEmbedAdvanced("Spotify", "Stopped adding results to the queue. A total of "+strconv.Itoa(page.AddedSoFar)+" tracks were added.", 0x1DB954)
		}

		page = nil
		return NewGenericEmbedAdvanced("Spotify", "Cancelled the Spotify session.", 0x1DB954)
	case "select", "choose", "play":
		if guildData[env.Guild.ID].SpotifyResults == nil {
			return NewErrorEmbed("Spotify Error", "No search session is in progress.")
		}
		if len(args) < 2 {
			return NewErrorEmbed("Spotify Error", "You must specify which result to select.")
		}

		page = guildData[env.Guild.ID].SpotifyResults[env.Message.Author.ID]
		results, _ := page.GetResults()

		switch args[1] {
		case "all", "*":
			foundVoiceChannel := false
			for _, voiceState := range env.Guild.VoiceStates {
				if voiceState.UserID == env.Message.Author.ID {
					foundVoiceChannel = true
					voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
					break
				}
			}
			if !foundVoiceChannel {
				return NewErrorEmbed("Spotify Error", "You must join the voice channel to use before using the "+args[0]+" command.")
			}
			for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
				//Wait for the handling of a previous playbck command to finish
			}

			//Update channel ID to send voice messages to
			guildData[env.Guild.ID].VoiceData.ChannelIDJoinedFrom = env.Channel.ID

			waitEmbed := NewEmbed().
				SetTitle("Spotify").
				SetDescription("Please wait a while as we add all " + strconv.Itoa(page.TotalResults) + " results to the queue...\n\nThe first result added will automatically begin playing. During this process, it may feel as if other commands are slow or don't work; give them some time to process.\nYou may cancel at any moment with ``" + env.BotPrefix + env.Command + " cancel``. Cancelling will not remove any results added to the queue during this process.").
				SetColor(0x1DB954).MessageEmbed
			botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, waitEmbed)

			page.AddingAll = true

			for i, result := range page.AllResults {
				//Give a chance for other commands waiting in line to execute
				guildData[env.Guild.ID].Unlock()
				guildData[env.Guild.ID].Lock()

				if page.Cancelled {
					page.AddingAll = false
					return nil
				}

				resultURL := "https://open.spotify.com/track/" + result.ID

				trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(result.URI)
				if err != nil {
					botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, NewErrorEmbed("Spotify Error", "Error finding track info for result "+strconv.Itoa(i+1)+"."))
					continue
				}

				queueData := AudioQueueEntry{MediaURL: resultURL, Requester: env.Message.Author, Type: "spotify",
					Author:       trackInfo.Artist,
					ImageURL:     trackInfo.ArtURL,
					ThumbnailURL: trackInfo.ArtURL,
					Title:        trackInfo.Title,
					Duration:     float64(trackInfo.Duration / 1000),
					TrackID:      trackInfo.TrackID,
				}

				if voiceIsStreaming(env.Guild.ID) {
					guildData[env.Guild.ID].QueueAdd(queueData)
					page.AddedSoFar++
					continue
				}

				guildData[env.Guild.ID].AudioNowPlaying = queueData
				page.AddedSoFar++
				go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)

				botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, queueData.GetNowPlayingEmbed())
			}

			page.AddingAll = false

			return NewGenericEmbedAdvanced("Spotify", "Finished adding all "+strconv.Itoa(page.TotalResults)+" results to the queue.", 0x1DB954)
		case "view":
			foundVoiceChannel := false
			for _, voiceState := range env.Guild.VoiceStates {
				if voiceState.UserID == env.Message.Author.ID {
					foundVoiceChannel = true
					voiceJoin(botData.DiscordSession, env.Guild.ID, voiceState.ChannelID, env.Message.ID)
					break
				}
			}
			if !foundVoiceChannel {
				return NewErrorEmbed("Spotify Error", "You must join the voice channel to use before using the "+args[0]+" command.")
			}
			for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
				//Wait for the handling of a previous playbck command to finish
			}

			//Update channel ID to send voice messages to
			guildData[env.Guild.ID].VoiceData.ChannelIDJoinedFrom = env.Channel.ID

			waitEmbed := NewEmbed().
				SetTitle("Spotify").
				SetDescription("Please wait a moment as we add all " + strconv.Itoa(len(page.Results)) + " results to the queue...\n\nThe first result added will automatically begin playing. During this process, it may feel as if other commands are slow or don't work; give them some time to process.\nYou may cancel at any moment with ``" + env.BotPrefix + env.Command + " cancel``. Cancelling will not remove any results added to the queue during this process.").
				SetColor(0x1DB954).MessageEmbed
			botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, waitEmbed)

			page.AddingAll = true

			for i, result := range page.Results {
				//Give a chance for other commands waiting in line to execute
				guildData[env.Guild.ID].Unlock()
				guildData[env.Guild.ID].Lock()

				if page.Cancelled {
					page.AddingAll = false
					return nil
				}

				resultURL := "https://open.spotify.com/track/" + result.ID

				trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(result.URI)
				if err != nil {
					botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, NewErrorEmbed("Spotify Error", "Error finding track info for result "+strconv.Itoa(i+1)+"."))
					continue
				}

				queueData := AudioQueueEntry{MediaURL: resultURL, Requester: env.Message.Author, Type: "spotify",
					Author:       trackInfo.Artist,
					ImageURL:     trackInfo.ArtURL,
					ThumbnailURL: trackInfo.ArtURL,
					Title:        trackInfo.Title,
					Duration:     float64(trackInfo.Duration / 1000),
					TrackID:      trackInfo.TrackID,
				}

				if voiceIsStreaming(env.Guild.ID) {
					guildData[env.Guild.ID].QueueAdd(queueData)
					page.AddedSoFar++
					continue
				}

				guildData[env.Guild.ID].AudioNowPlaying = queueData
				page.AddedSoFar++
				go voicePlayWrapper(botData.DiscordSession, env.Guild.ID, env.Channel.ID, queueData.MediaURL)

				botData.DiscordSession.ChannelMessageSendEmbed(env.Channel.ID, queueData.GetNowPlayingEmbed())
			}

			page.AddingAll = false

			return NewGenericEmbedAdvanced("Spotify", "Finished adding all "+strconv.Itoa(len(page.Results))+" results to the queue.", 0x1DB954)
		default:
			selection, err := strconv.Atoi(args[1])
			if err != nil {
				return NewErrorEmbed("Spotify Error", "``"+args[1]+"`` is not a valid number.")
			}
			if selection > len(results) || selection <= 0 {
				return NewErrorEmbed("Spotify Error", "An invalid selection was specified.")
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
				return NewErrorEmbed("Spotify Error", "You must join the voice channel to use before using the "+args[0]+" command.")
			}
			for guildData[env.Guild.ID].VoiceData.IsPlaybackPreparing {
				//Wait for the handling of a previous playback command to finish
			}

			//Update channel ID to send voice messages to
			guildData[env.Guild.ID].VoiceData.ChannelIDJoinedFrom = env.Channel.ID

			result := results[selection-1]
			resultURL := "https://open.spotify.com/track/" + result.ID

			queueData := AudioQueueEntry{MediaURL: resultURL, Requester: env.Message.Author, Type: "spotify"}
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
		}
	default:
		return NewErrorEmbed("Spotify Error", "Unknown command ``"+args[0]+"``.")
	}

	results, err := page.GetResults()
	if err != nil {
		return NewErrorEmbed("Spotify Error", "No search results were found.")
	}

	spotifyEmbed := NewEmbed().
		SetColor(0x1DB954)

	commandList := env.BotPrefix + env.Command + " play N - Plays result N" +
		"\n" + env.BotPrefix + env.Command + " cancel - Cancels the search session"
	if (page.PageNumber - 1) > 0 {
		commandList += "\n" + env.BotPrefix + env.Command + " prev - Displays the previous page"
	}
	if (page.PageNumber + 1) <= page.TotalPages {
		commandList += "\n" + env.BotPrefix + env.Command + " next - Displays the next page"
	}

	if page.IsPlaylist {
		spotifyEmbed.SetTitle("Spotify Playlist - Page " + strconv.Itoa(page.PageNumber)).
			SetDescription(strconv.Itoa(page.TotalResults) + " results for [" + page.Query + "](https://open.spotify.com/user/" + page.PlaylistUserID + "/playlist/" + page.PlaylistID + ")")
	} else {
		spotifyEmbed.SetTitle("Spotify Search Results - Page " + strconv.Itoa(page.PageNumber)).
			SetDescription(strconv.Itoa(page.TotalResults) + " results for \"" + page.Query + "\"")
	}

	commandListField := &discordgo.MessageEmbedField{Name: "Commands", Value: commandList}

	responseEmbed := spotifyEmbed.MessageEmbed

	fields := []*discordgo.MessageEmbedField{}
	for i := 0; i < len(results); i++ {
		trackInfo, err := botData.BotClients.Spotify.GetTrackInfo(results[i].URI)
		if err != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "Error fetching info for [this track](https://open.spotify.com/track/" + results[i].ID + ")"})
		} else {
			artist := trackInfo.Artist
			title := trackInfo.Title

			fields = append(fields, &discordgo.MessageEmbedField{Name: "Result #" + strconv.Itoa(i+1), Value: "[" + title + "](https://open.spotify.com/track/" + results[i].ID + ") by **" + artist + "**"})
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
				queueLength := len(guildData[env.Guild.ID].AudioQueue)

				guildData[env.Guild.ID].QueueClear()

				return NewGenericEmbed("Queue", "Cleared all "+strconv.Itoa(queueLength)+" entries from the queue.")
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

	pageNumber := 1
	if len(args) >= 1 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return NewErrorEmbed("Queue Error", "Invalid page number ``"+args[0]+"``.")
		}
		pageNumber = num
	}

	nowPlaying := guildData[env.Guild.ID].AudioNowPlaying
	nowPlayingField := &discordgo.MessageEmbedField{
		Name:  "Now Playing",
		Value: "There is no audio currently playing.",
	}

	if voiceIsStreaming(env.Guild.ID) && nowPlaying.MediaURL != "" {
		switch nowPlaying.Type {
		case "youtube":
			nowPlayingField.Name += " from YouTube"
			nowPlayingField.Value = fmt.Sprintf("[%s](%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
		case "soundcloud":
			nowPlayingField.Name += " from SoundCloud"
			nowPlayingField.Value = fmt.Sprintf("[%s](%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
		case "spotify":
			nowPlayingField.Name += " from Spotify"
			nowPlayingField.Value = fmt.Sprintf("[%s](https://open.spotify.com/track/%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
		default:
			if nowPlaying.Author != "" && nowPlaying.Title != "" {
				nowPlayingField.Value = fmt.Sprintf("[%s](%s) by **%s**\nRequested by <@!%s>", nowPlaying.Title, nowPlaying.MediaURL, nowPlaying.Author, nowPlaying.Requester.ID)
			} else {
				nowPlayingField.Value = fmt.Sprintf("%s\nRequested by <@!%s>", nowPlaying.MediaURL, nowPlaying.Requester.ID)
			}
		}
	}

	queueColors := make([]*colors.RGBColor, 0)
	queueList := make([]*discordgo.MessageEmbedField, 0)
	for queueEntryNumber, queueEntry := range guildData[env.Guild.ID].AudioQueue {
		displayNumber := strconv.Itoa(queueEntryNumber + 1)

		queueEntryFieldName := "Entry #" + displayNumber
		queueEntryFieldValue := ""

		if queueEntry.Title != "" && queueEntry.Author != "" {
			if queueEntry.Type == "spotify" {
				queueEntryFieldValue = "[" + queueEntry.Title + "](https://open.spotify.com/track/" + queueEntry.MediaURL + ") by **" + queueEntry.Author + "**"
			} else {
				queueEntryFieldValue = "[" + queueEntry.Title + "](" + queueEntry.MediaURL + ") by **" + queueEntry.Author + "**"
			}
		} else {
			queueEntryFieldValue = queueEntry.MediaURL
		}
		queueEntryFieldValue += "\nRequested by <@!" + queueEntry.Requester.ID + ">"

		queueList = append(queueList, &discordgo.MessageEmbedField{Name: queueEntryFieldName, Value: queueEntryFieldValue})

		switch queueEntry.Type {
		case "youtube":
			color, _ := colors.RGB(0xFF, 0x00, 0x00)
			queueColors = append(queueColors, color)
		case "soundcloud":
			color, _ := colors.RGB(0xFF, 0x77, 0x00)
			queueColors = append(queueColors, color)
		case "spotify":
			color, _ := colors.RGB(0x1D, 0xB9, 0x54)
			queueColors = append(queueColors, color)
		default:
			color, _ := colors.RGB(0x1C, 0x1C, 0x1C)
			queueColors = append(queueColors, color)
		}
	}

	if len(queueList) <= 0 {
		queueEmbed := NewEmbed().
			SetTitle("Queue Error").
			SetDescription("No queue entries found.").
			SetThumbnail(nowPlaying.ThumbnailURL)

		queueEmbed.Fields = append(queueEmbed.Fields, nowPlayingField)

		return queueEmbed.MessageEmbed
	}

	pagedQueueList, totalPages, err := page(queueList, pageNumber, 10)
	if err != nil {
		return NewErrorEmbed("Queue Error", fmt.Sprintf("%v", err))
	}

	queueColor := 0x1C1C1C
	if len(queueColors) > 0 {
		queueColorR := uint8(0x0)
		queueColorG := uint8(0x0)
		queueColorB := uint8(0x0)
		for _, color := range queueColors {
			queueColorR += color.R
			queueColorG += color.G
			queueColorB += color.B
		}
		queueColorsLen := uint8(len(queueColors))
		queueColorR /= queueColorsLen
		queueColorG /= queueColorsLen
		queueColorB /= queueColorsLen
		queueColor = int(queueColorR + queueColorG + queueColorB)
	}

	queueEmbed := pagedQueueList.
		SetTitle("Queue for " + env.Guild.Name + " - Page " + strconv.Itoa(pageNumber) + "/" + strconv.Itoa(totalPages)).
		SetDescription("There are " + strconv.Itoa(len(queueList)) + " entries in the queue.").
		SetColor(queueColor).
		SetThumbnail(nowPlaying.ThumbnailURL)

	queueEmbed.Fields = append(queueEmbed.Fields, nowPlayingField)

	return queueEmbed.MessageEmbed
}

func commandNowPlaying(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if voiceIsStreaming(env.Guild.ID) {
		return guildData[env.Guild.ID].AudioNowPlaying.GetNowPlayingDurationEmbed(guildData[env.Guild.ID].VoiceData.StreamingSession)
	}
	return NewErrorEmbed("Now Playing Error", "There is no audio currently playing.")
}
