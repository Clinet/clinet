package main

import (
	"encoding/json"
	"sort"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// Starboard holds data specific to a guild's starboard
type Starboard struct {
	Active            bool             //Whether or not the starboard is active
	AllowSelfStar     bool             //Whether or not a user may star their own message to add it to the starboard
	BlacklistChannels []string         //A list of channel IDs to exclude from the starboard
	BlacklistUsers    []string         //A list of user IDs to exclude from the starboard
	ChannelID         string           //The channel to use as a starboard
	Emoji             string           //The emoji to use for the starboard
	NSFWChannelID     string           //The NSFW channel to use as a starboard for NSFW channels
	NSFWEmoji         string           //The emoji to use for the NSFW starboard
	MinimumStars      int              //The minimum amount of stars that must exist for a starboard entry to be made
	StarboardEntries  []StarboardEntry //A list of starboard entries in a string map, where key = reaction message ID and value = starboard entry message ID
}

// StarboardEntry holds data specific to a starboard entry
// The only two things *discordgo.Session.ChannelMessage() needs are the channel IDs and the message IDs
// Mistakes were made in early attempts by not storing the channel IDs, resulting in being unable to check messages for reaction updates without an event being triggered
type StarboardEntry struct {
	SourceChannelID    string //The channel ID the source message resides in
	SourceMessageID    string //The source message ID
	StarboardChannelID string //The channel ID the starboard entry message resides in (just in case the starboard channel changes and messages need to be moved to a new channel)
	StarboardMessageID string //The starboard entry's message ID
	Stars              int    //The amount of stars on this entry
}

func commandStarboard(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "debug":
		if env.User.ID != botData.BotOwnerID {
			NewErrorEmbed("Command Error - Not Authorized (NA)", "You are not authorized to use this command.")
		}
		starboard := starboards[env.Guild.ID]
		json, _ := json.MarshalIndent(starboard, "", "")
		return NewGenericEmbed("Starboard - Debug", string(json))
	case "stats":
		//Go through starboard for this guild and only pull entries from the caller
		//Build list in embed
		//Return message to user

		return nil
	case "minimum":
		if len(args) == 1 {
			if env.Channel.NSFW {
				return NewGenericEmbed("Starboard", "Minimum required "+starboards[env.Guild.ID].NSFWEmoji+" reactions: "+strconv.Itoa(starboards[env.Guild.ID].MinimumStars))
			}
			return NewGenericEmbed("Starboard", "Minimum required "+starboards[env.Guild.ID].Emoji+" reactions: "+strconv.Itoa(starboards[env.Guild.ID].MinimumStars))
		}

		minimum, err := strconv.Atoi(args[1])
		if err != nil {
			return NewErrorEmbed("Starboard Error", "``"+args[1]+"`` is not a valid number.")
		}

		starboards[env.Guild.ID].MinimumStars = minimum
		return NewGenericEmbed("Starboard", "Successfully set the minimum required reactions to "+args[1]+".")
	case "leaderboard":
		if len(args) == 1 {
			//Go through starboard for this guild
			starboardEntries := starboards[env.Guild.ID].StarboardEntries

			//Remove NSFW starboard entries from leaderboard if not NSFW channel
			if env.Channel.NSFW == false {
				for i, starboardEntry := range starboardEntries {
					if starboardEntry.SourceChannelID == starboards[env.Guild.ID].NSFWChannelID {
						starboardEntries = append(starboards[env.Guild.ID].StarboardEntries[:i], starboards[env.Guild.ID].StarboardEntries[i+1])
						i--
					}
				}
			}

			//Build list in order from most stars to least
			sort.Slice(starboardEntries, func(i, j int) bool { return starboardEntries[i].Stars < starboardEntries[j].Stars })

			//Generate embed for leaderboard
			leaderboardEmbed := NewEmbed().
				SetAuthor(env.Guild.Name, "https://cdn.discordapp.com/icons/"+env.Guild.ID+"/"+env.Guild.Icon+".jpg").
				SetTitle("Starboard - Leaderboard").
				SetDescription("__Top 10 Starboard Entries__").
				SetColor(0xFFE200)

			for i, starboardEntry := range starboardEntries {
				sourceMessage, err := botData.DiscordSession.ChannelMessage(starboardEntry.SourceChannelID, starboardEntry.SourceMessageID)
				if err != nil {
					starboardEntries = append(starboards[env.Guild.ID].StarboardEntries[:i], starboards[env.Guild.ID].StarboardEntries[i+1])
					i--
					continue
				}
				sourceChannel, err := botData.DiscordSession.Channel(starboardEntry.SourceChannelID)
				if err != nil {
					starboardEntries = append(starboards[env.Guild.ID].StarboardEntries[:i], starboards[env.Guild.ID].StarboardEntries[i+1])
					i--
					continue
				}
				leaderboardEmbed.AddField(starboards[env.Guild.ID].Emoji+" "+strconv.Itoa(starboardEntry.Stars)+" - "+sourceMessage.Author.Username+"#"+sourceMessage.Author.Discriminator+" in #"+sourceChannel.Name, sourceMessage.Content)
			}

			//Return message to user
			return leaderboardEmbed.MessageEmbed
		}

		//Find first mentioned user
		//Go through starboard for this guild and only pull entries from mentioned user
		//Build list in embed
		//Return message to user

		return nil
	case "enable":
		starboards[env.Guild.ID].Active = true
		return NewGenericEmbed("Starboard", "Enabled the starboard successfully.")
	case "disable":
		starboards[env.Guild.ID].Active = false
		return NewGenericEmbed("Starboard", "Disabled the starboard successfully.")
	case "channel":
		if len(args) == 1 {
			if starboards[env.Guild.ID].ChannelID == "" {
				return NewGenericEmbed("Starboard", "No starbard channel has been set.")
			}
			return NewGenericEmbed("Starboard", "Starboad channel: <#"+starboards[env.Guild.ID].ChannelID+">")
		}
		if args[1] == "set" {
			starboards[env.Guild.ID].ChannelID = env.Channel.ID
			return NewGenericEmbed("Starboard", "Set the starboard channel to <#"+env.Channel.ID+">.")
		}
		if args[1] == "remove" {
			starboards[env.Guild.ID].ChannelID = ""
			return NewGenericEmbed("Starboard", "Unset the previous starboard channel.")
		}
		return NewErrorEmbed("Starboard Error", "You must specify ``set`` instead of ``"+args[1]+"`` to set the current channel as the starboard channel.")
	case "nsfwchannel":
		if len(args) == 1 {
			if starboards[env.Guild.ID].NSFWChannelID == "" {
				return NewGenericEmbed("Starboard", "No NSFW starbard channel has been set.")
			}
			return NewGenericEmbed("Starboard", "NSFW starboad channel: <#"+starboards[env.Guild.ID].NSFWChannelID+">")
		}
		if args[1] == "set" {
			if !env.Channel.NSFW {
				return NewErrorEmbed("Starboard Error", "You must mark this channel as NSFW before you can use it as the NSFW starboard channel.")
			}
			starboards[env.Guild.ID].NSFWChannelID = env.Channel.ID
			return NewGenericEmbed("Starboard", "Set the NSFW starboard channel to <#"+env.Channel.ID+">.")
		}
		if args[1] == "remove" {
			starboards[env.Guild.ID].NSFWChannelID = ""
			return NewGenericEmbed("Starboard", "Unset the previous NSFW starboard channel.")
		}
		return NewErrorEmbed("Starboard Error", "You must specify ``set`` instead of ``"+args[1]+"`` to set the current channel as the NSFW starboard channel.")
	case "emoji":
		if len(args) == 1 {
			return NewGenericEmbed("Starboard", "Emoji: "+starboards[env.Guild.ID].Emoji)
		}
		starboards[env.Guild.ID].Emoji = args[1]
		return NewGenericEmbed("Starboard", "Set the emoji to "+args[1]+".")
	case "nsfwemoji":
		if len(args) == 1 {
			return NewGenericEmbed("Starboard", "NSFW Emoji: "+starboards[env.Guild.ID].NSFWEmoji)
		}
		starboards[env.Guild.ID].Emoji = args[1]
		return NewGenericEmbed("Starboard", "Set the NSFW emoji to "+args[1]+".")
	case "selfstar":
		if len(args) == 1 {
			return NewGenericEmbed("Starboard", "Allow selfstar: **"+strconv.FormatBool(starboards[env.Guild.ID].AllowSelfStar)+"**")
		}
		switch args[1] {
		case "true", "yes", "enable":
			starboards[env.Guild.ID].AllowSelfStar = true

			//Apparently Discord doesn't send enough info in the reactions object of a message
			//I'll build up a list of who reacted with what later on in life, too much for now so selfstars won't get added for now

			return NewGenericEmbed("Starboard", "Successfully enabled selfstar.")
		case "false", "no", "disable":
			starboards[env.Guild.ID].AllowSelfStar = false

			//Apparently Discord doesn't send enough info in the reactions object of a message
			//I'll build up a list of who reacted with what later on in life, too much for now so selfstars won't get removed for now

			return NewGenericEmbed("Starboard", "Successfully disabled selfstar.")
		default:
			return NewErrorEmbed("Starboard Error", "Unknown value ``"+args[1]+"``. Please use either ``enable`` or ``disable``.")
		}
	}
	return NewErrorEmbed("Starboard Error", "Error finding the setting ``"+args[0]+"``.")
}

func discordMessageReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	channel, err := session.Channel(reaction.ChannelID)
	if err != nil {
		return
	}

	if _, exists := starboards[channel.GuildID]; exists == false {
		return
	}

	if starboards[channel.GuildID].Active == false {
		return
	}
	if starboards[channel.GuildID].NSFWChannelID == "" && starboards[channel.GuildID].ChannelID == "" {
		return
	}
	if channel.NSFW && starboards[channel.GuildID].NSFWChannelID == "" {
		return
	}
	if channel.NSFW == false && starboards[channel.GuildID].ChannelID == "" {
		return
	}

	message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}
	if message.Author.ID == session.State.User.ID {
		return
	}

	//A user can't self-star their message to add it to the starboard, however I give up on finding
	//a method of subtracting their star for now so it'll still show the total star count
	if message.Author.ID == reaction.UserID && starboards[channel.GuildID].AllowSelfStar == false {
		return
	}

	stars := 0
	for _, msgReaction := range message.Reactions {
		if channel.NSFW {
			if msgReaction.Emoji.Name == starboards[channel.GuildID].NSFWEmoji {
				stars = msgReaction.Count
				break
			}
			if msgReaction.Emoji.ID == starboards[channel.GuildID].NSFWEmoji {
				stars = msgReaction.Count
				break
			}
		} else {
			if msgReaction.Emoji.Name == starboards[channel.GuildID].Emoji {
				stars = msgReaction.Count
				break
			}
			if msgReaction.Emoji.ID == starboards[channel.GuildID].Emoji {
				stars = msgReaction.Count
				break
			}
		}
	}
	if stars == 0 {
		return
	}
	if stars < starboards[channel.GuildID].MinimumStars {
		return
	}

	entry := createStarboardEntry(stars, message, channel)

	//Check to see if the entry already exists, and if so, update it instead of create a new one
	for _, starboardEntry := range starboards[channel.GuildID].StarboardEntries {
		if starboardEntry.SourceMessageID == message.ID {
			if channel.NSFW {
				session.ChannelMessageEditEmbed(starboards[channel.GuildID].NSFWChannelID, starboardEntry.SourceMessageID, entry)
			} else {
				session.ChannelMessageEditEmbed(starboards[channel.GuildID].ChannelID, starboardEntry.SourceMessageID, entry)
			}
			return
		}
	}

	//Create a new entry
	if channel.NSFW {
		starboardMessage, err := session.ChannelMessageSendEmbed(starboards[channel.GuildID].NSFWChannelID, entry)
		if err != nil {
			return
		}

		starboards[channel.GuildID].StarboardEntries = append(starboards[channel.GuildID].StarboardEntries, StarboardEntry{
			SourceChannelID:    channel.ID,
			SourceMessageID:    message.ID,
			StarboardChannelID: starboards[channel.GuildID].ChannelID,
			StarboardMessageID: starboardMessage.ID,
			Stars:              stars,
		})
	} else {
		starboardMessage, err := session.ChannelMessageSendEmbed(starboards[channel.GuildID].ChannelID, entry)
		if err != nil {
			return
		}

		starboards[channel.GuildID].StarboardEntries = append(starboards[channel.GuildID].StarboardEntries, StarboardEntry{
			SourceChannelID:    channel.ID,
			SourceMessageID:    message.ID,
			StarboardChannelID: starboards[channel.GuildID].ChannelID,
			StarboardMessageID: starboardMessage.ID,
			Stars:              stars,
		})
	}
}
func discordMessageReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	channel, err := session.Channel(reaction.ChannelID)
	if err != nil {
		return
	}

	if _, exists := starboards[channel.GuildID]; !exists {
		return
	}

	if starboards[channel.GuildID].Active == false {
		return
	}
	if starboards[channel.GuildID].NSFWChannelID == "" && starboards[channel.GuildID].ChannelID == "" {
		return
	}
	if channel.NSFW && starboards[channel.GuildID].NSFWChannelID == "" {
		return
	}
	if !channel.NSFW && starboards[channel.GuildID].ChannelID == "" {
		return
	}

	message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}

	stars := 0
	for _, msgReaction := range message.Reactions {
		if channel.NSFW {
			if msgReaction.Emoji.Name == starboards[channel.GuildID].NSFWEmoji {
				stars = msgReaction.Count
				break
			}
			if msgReaction.Emoji.ID == starboards[channel.GuildID].NSFWEmoji {
				stars = msgReaction.Count
				break
			}
		} else {
			if msgReaction.Emoji.Name == starboards[channel.GuildID].Emoji {
				stars = msgReaction.Count
				break
			}
			if msgReaction.Emoji.ID == starboards[channel.GuildID].Emoji {
				stars = msgReaction.Count
				break
			}
		}
	}
	if stars == 0 || stars < starboards[channel.GuildID].MinimumStars {
		for i, starboardEntry := range starboards[channel.GuildID].StarboardEntries {
			if starboardEntry.SourceMessageID == message.ID {
				if channel.NSFW {
					session.ChannelMessageDelete(starboards[channel.GuildID].NSFWChannelID, starboardEntry.StarboardMessageID)
				} else {
					session.ChannelMessageDelete(starboards[channel.GuildID].ChannelID, starboardEntry.StarboardMessageID)
				}
				starboards[channel.GuildID].StarboardEntries = append(starboards[channel.GuildID].StarboardEntries[:i], starboards[channel.GuildID].StarboardEntries[i+1])
				return
			}
		}
		return
	}

	entry := createStarboardEntry(stars, message, channel)

	//Check to see if the entry already exists, and if so, update it instead of create a new one
	for i, starboardEntry := range starboards[channel.GuildID].StarboardEntries {
		if starboardEntry.SourceMessageID == message.ID {
			if channel.NSFW {
				session.ChannelMessageEditEmbed(starboards[channel.GuildID].NSFWChannelID, starboardEntry.StarboardMessageID, entry)
				starboards[channel.GuildID].StarboardEntries[i].Stars = stars
			} else {
				session.ChannelMessageEditEmbed(starboards[channel.GuildID].ChannelID, starboardEntry.StarboardMessageID, entry)
				starboards[channel.GuildID].StarboardEntries[i].Stars = stars
			}
			return
		}
	}

	//Create a new entry
	if channel.NSFW {
		starboardMessage, err := session.ChannelMessageSendEmbed(starboards[channel.GuildID].NSFWChannelID, entry)
		if err != nil {
			return
		}

		starboards[channel.GuildID].StarboardEntries = append(starboards[channel.GuildID].StarboardEntries, StarboardEntry{
			SourceChannelID:    channel.ID,
			SourceMessageID:    message.ID,
			StarboardChannelID: starboards[channel.GuildID].ChannelID,
			StarboardMessageID: starboardMessage.ID,
		})
	} else {
		starboardMessage, err := session.ChannelMessageSendEmbed(starboards[channel.GuildID].ChannelID, entry)
		if err != nil {
			return
		}

		starboards[channel.GuildID].StarboardEntries = append(starboards[channel.GuildID].StarboardEntries, StarboardEntry{
			SourceChannelID:    channel.ID,
			SourceMessageID:    message.ID,
			StarboardChannelID: starboards[channel.GuildID].ChannelID,
			StarboardMessageID: starboardMessage.ID,
		})
	}
}
func discordMessageReactionRemoveAll(session *discordgo.Session, reaction *discordgo.MessageReactionRemoveAll) {
	channel, err := session.Channel(reaction.ChannelID)
	if err != nil {
		return
	}

	if _, exists := starboards[channel.GuildID]; !exists {
		return
	}

	if starboards[channel.GuildID].Active == false {
		return
	}
	if starboards[channel.GuildID].NSFWChannelID == "" && starboards[channel.GuildID].ChannelID == "" {
		return
	}
	if channel.NSFW && starboards[channel.GuildID].NSFWChannelID == "" {
		return
	}
	if !channel.NSFW && starboards[channel.GuildID].ChannelID == "" {
		return
	}

	message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
	if err != nil {
		return
	}

	for i, starboardEntry := range starboards[channel.GuildID].StarboardEntries {
		if starboardEntry.SourceMessageID == message.ID {
			if channel.NSFW {
				session.ChannelMessageDelete(starboards[channel.GuildID].NSFWChannelID, starboardEntry.StarboardMessageID)
			} else {
				session.ChannelMessageDelete(starboards[channel.GuildID].ChannelID, starboardEntry.StarboardMessageID)
			}
			starboards[channel.GuildID].StarboardEntries = append(starboards[channel.GuildID].StarboardEntries[:i], starboards[channel.GuildID].StarboardEntries[i+1])
			return
		}
	}
}

func createStarboardEntry(stars int, message *discordgo.Message, channel *discordgo.Channel) *discordgo.MessageEmbed {
	entry := NewEmbed().
		SetAuthor(message.Author.Username+"#"+message.Author.Discriminator+" in #"+channel.Name, message.Author.AvatarURL("2048"))

	if channel.NSFW {
		entry.SetFooter(starboards[channel.GuildID].NSFWEmoji + " " + strconv.Itoa(stars)).
			SetColor(0xDEA7DF)
	} else {
		entry.SetFooter(starboards[channel.GuildID].Emoji + " " + strconv.Itoa(stars)).
			SetColor(0xFFE200)
	}

	if message.Content != "" {
		entry.SetDescription(message.Content)
	}
	if len(message.Attachments) > 0 {
		for _, attachment := range message.Attachments {
			if attachment.Width > 0 && attachment.Height > 0 {
				entry.SetImage(attachment.URL)
				break
			}
		}
	}

	return entry.MessageEmbed
}
