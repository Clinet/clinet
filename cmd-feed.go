package main

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

var (
	regexParagraph = regexp.MustCompile("(?s)<p>(.*)</p>")
	regexLink      = regexp.MustCompile("(?s)<a href=\"(.*)\">(.*)</a>")
	regexCode      = regexp.MustCompile("(?s)<code(?:.*)>(.*)</code>")
)

// A wrapper for *gofeed.Feed
type Feed struct {
	*gofeed.Feed

	ChannelID string `json:"channelID"` //The channel to post new feed entries to
	FeedURL   string `json:"feedURL"`   //The URL to the feed
	Frequency int    `json:"frequency"` //How often to check for new feed entries in seconds
}

func commandFeed(args []CommandArgument, env *CommandEnvironment) *discordgo.MessageEmbed {
	feedsToAdd := make([]string, 0)
	isAdding := false

	feedsToEdit := make([]int, 0)
	isEditing := false

	feedsToRemove := make([]int, 0)
	isRemoving := false

	isListing := false

	//isAll := false
	isSettingChannel := false
	frequency := botData.BotOptions.FeedFrequency

	for _, arg := range args {
		switch arg.Name {
		case "add":
			if len(feedsToEdit) > 0 || len(feedsToRemove) > 0 || isListing {
				return NewErrorEmbed("Feed Error", "You cannot mix arguments dictating what to do!")
			}
			if arg.Value == "" {
				return NewErrorEmbed("Feed Error", "You must specify a feed to add when using the ``-add`` argument.")
			}
			if _, err := url.ParseRequestURI(arg.Value); err != nil {
				return NewErrorEmbed("Feed Error", "``"+arg.Value+"`` is not a valid URL.")
			}

			for _, feed := range guildSettings[env.Guild.ID].Feeds {
				if arg.Value == feed.FeedLink {
					return NewErrorEmbed("Feed Error", "Feed ``"+arg.Value+"`` already exists.")
				}
			}

			isAdding = true
			feedsToAdd = append(feedsToAdd, arg.Value)
		case "frequency", "f":
			if arg.Value == "" {
				return NewErrorEmbed("Feed Error", "You must specify a post check frequency to use when using the ``-"+arg.Name+"`` argument.")
			}
			freq, err := strconv.Atoi(arg.Value)
			if err != nil {
				return NewErrorEmbed("Feed Error", "``"+arg.Value+"`` is not a valid number.")
			}
			if freq < botData.BotOptions.FeedFrequency {
				return NewErrorEmbed("Feed Error", "Frequency must not be lower than "+strconv.Itoa(botData.BotOptions.FeedFrequency)+" seconds.")
			}
			frequency = freq
			//		case "all":
			//			isAll = true
		case "list":
			if len(feedsToAdd) > 0 || len(feedsToEdit) > 0 || len(feedsToRemove) > 0 {
				return NewErrorEmbed("Feed Error", "You cannot mix arguments dictating what to do!")
			}
			isListing = true
		case "setchannel":
			isSettingChannel = true
		case "edit":
			if len(feedsToAdd) > 0 || len(feedsToRemove) > 0 || isListing {
				return NewErrorEmbed("Feed Error", "You cannot mix arguments dictating what to do!")
			}
			if arg.Value == "" {
				return NewErrorEmbed("Feed Error", "You must specify a feed entry to edit when using the ``-edit`` argument.")
			}
			entry, err := strconv.Atoi(arg.Value)
			if err != nil {
				return NewErrorEmbed("Feed Error", "``"+arg.Value+"`` is not a valid number.")
			}
			if entry > len(guildSettings[env.Guild.ID].Feeds) || entry <= 0 {
				return NewErrorEmbed("Feed Error", "``"+arg.Value+"`` is not a valid feed entry.")
			}

			isEditing = true
			feedsToEdit = append(feedsToEdit, entry)
		case "remove":
			if len(feedsToAdd) > 0 || len(feedsToEdit) > 0 || isListing {
				return NewErrorEmbed("Feed Error", "You cannot mix arguments dictating what to do!")
			}
			if arg.Value == "" {
				return NewErrorEmbed("Feed Error", "You must specify a feed entry to remove when using the ``-remove`` argument.")
			}
			entry, err := strconv.Atoi(arg.Value)
			if err != nil {
				return NewErrorEmbed("Feed Error", "``"+arg.Value+"`` is not a valid number.")
			}
			if entry > len(guildSettings[env.Guild.ID].Feeds) || entry <= 0 {
				return NewErrorEmbed("Feed Error", "``"+arg.Value+"`` is not a valid feed entry.")
			}

			isRemoving = true
			feedsToRemove = append(feedsToRemove, entry)
		}
	}

	if isListing {
		if len(guildSettings[env.Guild.ID].Feeds) == 0 {
			return NewGenericEmbed("Feed", "There are no feed entries to list!")
		}

		feedListEmbed := NewEmbed().
			SetTitle("Feed Entries").
			SetDescription("A list of all available feed entries in this server.").
			SetColor(0x1C1C1C)

		for i, feedEntry := range guildSettings[env.Guild.ID].Feeds {
			feedListEmbed.AddField("Entry #"+strconv.Itoa(i+1), "Channel: <#"+feedEntry.ChannelID+">\nFeed: "+feedEntry.FeedLink)
		}

		return feedListEmbed.MessageEmbed
	}
	if isAdding {
		failedAdds := make([]string, 0)

		for _, feedURL := range feedsToAdd {
			addErr := addFeed(env.Guild.ID, env.Channel.ID, feedURL, frequency)
			if addErr != nil {
				failedAdds = append(failedAdds, feedURL)
			}
		}

		if len(failedAdds) > 0 {
			return NewGenericEmbed("Feed", "Some feed entries may have been added successfully, but the following feed entries failed to be processed: \n- "+strings.Join(failedAdds, "\n- "))
		}

		return NewGenericEmbed("Feed", "Successfully added the specified feed entries.")
	}
	if isEditing {
		for _, feedEntry := range feedsToEdit {
			guildSettings[env.Guild.ID].Feeds[feedEntry-1].Frequency = frequency
			if isSettingChannel {
				guildSettings[env.Guild.ID].Feeds[feedEntry-1].ChannelID = env.Channel.ID
			}
		}

		return NewGenericEmbed("Feed", "Successfully modified the specified feed entries.")
	}
	if isRemoving {
		newFeeds := make([]*Feed, 0)

		for i, feedEntry := range guildSettings[env.Guild.ID].Feeds {
			removed := false

			for _, removedEntry := range feedsToRemove {
				if i == (removedEntry - 1) {
					removed = true
					break
				}
			}

			if removed == false {
				newFeeds = append(newFeeds, feedEntry)
			}
		}

		guildSettings[env.Guild.ID].Feeds = newFeeds

		return NewGenericEmbed("Feed", "Successfully removed the specified feed entries.")
	}

	return nil
}

func addFeed(guildID, channelID, feedURL string, frequency int) error {
	feed, err := botData.BotClients.FeedParser.ParseURL(feedURL)
	if err != nil {
		return err
	}

	wrapFeed := &Feed{Feed: feed}
	wrapFeed.ChannelID = channelID
	wrapFeed.FeedURL = feedURL
	wrapFeed.Frequency = frequency

	guildSettings[guildID].Feeds = append(guildSettings[guildID].Feeds, wrapFeed)

	waitDuration := time.Duration(frequency) * time.Second
	time.AfterFunc(waitDuration, func() {
		postFeed(guildID, len(guildSettings[guildID].Feeds)-1, wrapFeed.Title, frequency)
	})

	return nil
}

// Takes both a pointer to an entry in the feed list and a feedURL to compare it against, posting new feed entries if found
//
// If the comparison succeeds, a post check will be made, and a new post will be posted if found.
// In this case, the postFeed function will be re-registered for a later call.
//
// If the comparison fails, it means that the given feedPointer no longer points to its original feed as the original feed was removed.
// In this case, the postFeed function will not be re-registered for a later call.
func postFeed(guildID string, feedPointer int, feedTitle string, frequency int) {
	if len(guildSettings[guildID].Feeds) == 0 {
		return
	}
	if feedPointer < 0 {
		return
	}
	if feedPointer >= len(guildSettings) {
		return
	}

	feed := guildSettings[guildID].Feeds[feedPointer]
	if feed.Title != feedTitle {
		return
	}

	waitDuration := time.Duration(frequency) * time.Second
	time.AfterFunc(waitDuration, func() {
		postFeed(guildID, feedPointer, feed.Title, frequency)
	})

	newFeed, err := botData.BotClients.FeedParser.ParseURL(feed.FeedURL)
	if err != nil {
		return
	}

	newPostCount := 0
	for _, newPost := range newFeed.Items {
		if newPost.GUID == feed.Items[0].GUID {
			if newPost.Updated == feed.Items[0].Updated {
				break
			}
		}
		if newPost.Title == feed.Items[0].Title {
			if newPost.Updated == feed.Items[0].Updated {
				break
			}
		}
		newPostCount++
	}

	if newPostCount > 0 {
		newPosts := newFeed.Items[0:newPostCount]

		for _, post := range newPosts {
			content := post.Content
			content = regexParagraph.ReplaceAllString(content, "$1\n\n")
			content = regexLink.ReplaceAllString(content, "[$1]($2)")
			content = regexCode.ReplaceAllString(content, "``$1``")

			postEmbed := NewEmbed().
				SetTitle(newFeed.Title).
				SetDescription(post.Link).
				AddField(post.Title, content).
				SetFooter("Updated " + post.Updated).
				SetColor(0x1C1C1C).MessageEmbed
			botData.DiscordSession.ChannelMessageSendEmbed(feed.ChannelID, postEmbed)
		}

		wrapFeed := &Feed{Feed: newFeed}
		wrapFeed.ChannelID = feed.ChannelID
		wrapFeed.FeedURL = feed.FeedURL
		wrapFeed.Frequency = feed.Frequency

		guildSettings[guildID].Feeds[feedPointer] = wrapFeed
	}
}
