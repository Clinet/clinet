package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"4d63.com/tz"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/olebedev/when"
)

// RemindEntry stores information about a remind entry
type RemindEntry struct {
	UserID    string    `json:"userID"`
	ChannelID string    `json:"channelID"`
	GuildID   string    `json:"guildID"`
	Message   string    `json:"message"`
	Added     time.Time `json:"timeAdded"`
	When      time.Time `json:"timeRemind"`
}

func commandRemind(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	timezone := userSettings[env.User.ID].Timezone
	if timezone == "" {
		return NewErrorEmbed("Remind Error", "Please set a timezone first!\n\nEx: ``"+env.BotPrefix+"user timezone America/New_York``")
	}
	location, err := tz.LoadLocation(timezone)
	if err != nil {
		return NewErrorEmbed("Remind Error", "You have an invalid timezone set, please set a new one first!\n\nEx: ``"+env.BotPrefix+"user timezone America/New_York``")
	}

	switch args[0] {
	case "list":
		pageNumber := 1
		if len(args) == 2 {
			page, err := strconv.Atoi(args[1])
			if err != nil {
				return NewErrorEmbed("Remind Error", "Invalid page number ``"+args[0]+"``.")
			}
			pageNumber = page
		}

		remindList := make([]*discordgo.MessageEmbedField, 0)
		for _, entry := range remindEntries {
			if entry.UserID == env.User.ID {
				remindList = append(remindList, &discordgo.MessageEmbedField{
					Name:  "Entry #" + strconv.Itoa(len(remindList)+1) + " - " + entry.When.String(),
					Value: entry.Message,
				})
			}
		}

		remindListEmbed, totalPages, err := page(remindList, pageNumber, 10)
		if totalPages == 0 {
			return NewGenericEmbed("Remind", "No remind entries were found.")
		}
		if err != nil {
			return NewErrorEmbed("Remind Error", "Invalid page number ``"+strconv.Itoa(pageNumber)+"``.")
		}

		return remindListEmbed.SetTitle("Remind List - Page " + strconv.Itoa(pageNumber) + "/" + strconv.Itoa(totalPages)).MessageEmbed
	case "delete", "remove":
		remindList := make([]RemindEntry, 0)
		for _, entry := range remindEntries {
			if entry.UserID == env.User.ID {
				remindList = append(remindList, entry)
			}
		}

		debugLog(fmt.Sprintf("%v", remindList), true)

		for _, remindEntry := range args[1:] {
			remindEntryNumber, err := strconv.Atoi(remindEntry)
			if err != nil {
				return NewErrorEmbed("Remind Error", "``"+remindEntry+"`` is not a valid number.")
			}
			remindEntryNumber--

			if remindEntryNumber >= len(remindList) || remindEntryNumber < 0 {
				return NewErrorEmbed("Remind Error", "``"+remindEntry+"`` is not a valid remind entry.")
			}
		}

		var newRemindList []RemindEntry
		for remindEntryN, remindEntry := range remindList {
			keepRemindEntry := true
			for _, removedRemindEntry := range args[1:] {
				removedRemindEntryNumber, _ := strconv.Atoi(removedRemindEntry)
				removedRemindEntryNumber--
				if remindEntryN == removedRemindEntryNumber {
					keepRemindEntry = false
					break
				}
			}
			if keepRemindEntry {
				newRemindList = append(newRemindList, remindEntry)
			}
		}

		debugLog(fmt.Sprintf("%v", newRemindList), true)

		var newRemindEntries []RemindEntry
		for _, remindEntry := range remindEntries {
			if remindEntry.UserID != env.User.ID {
				newRemindEntries = append(newRemindEntries, remindEntry)
				continue
			}
			keepRemindEntry := false
			for _, remindEntryKeep := range newRemindList {
				if remindEntry.ChannelID == remindEntryKeep.ChannelID && remindEntry.Message == remindEntryKeep.Message {
					keepRemindEntry = true
					break
				}
			}
			if keepRemindEntry {
				newRemindEntries = append(newRemindEntries, remindEntry)
			}
		}

		debugLog(fmt.Sprintf("%v", newRemindEntries), true)

		remindEntries = newRemindEntries

		if len(args) > 2 {
			return NewGenericEmbed("Remind", "Successfully removed the specified remind entries.")
		}
		return NewGenericEmbed("Remind", "Successfully removed the specified remind entry.")
	}

	w := when.EN
	text := strings.Join(args, " ")
	now := time.Now().In(location)

	r, err := w.Parse(text, now)
	if err != nil || r == nil {
		return NewErrorEmbed("Remind Error", "There was an error figuring out what time to remind you with this message at.")
	}

	waitDuration := r.Time.In(location).Sub(now)
	if waitDuration < 0 {
		return NewErrorEmbed("Remind Error", "That time was "+humanize.Time(r.Time.In(location))+"!")
	}

	defer remindWhen(env.User.ID, env.Guild.ID, env.Channel.ID, text, now.In(location), r.Time.In(location), now.In(location))

	return NewEmbed().
		SetTitle("Remind").
		SetDescription("I will give you this reminder "+humanize.Time(r.Time.In(location))+" at ``"+r.Time.In(location).String()+"``.").
		AddField("Reminder", r.Source).
		SetColor(0x1C1C1C).MessageEmbed
}

func remindWhen(userID, guildID, channelID, message string, added, when, now time.Time) {
	remindEntries = append(remindEntries, RemindEntry{UserID: userID, ChannelID: channelID, Message: message, Added: added, When: when})

	waitDuration := when.Sub(now)
	time.AfterFunc(waitDuration, func() {
		botData.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content: "<@!" + userID + "> :alarm_clock:",
			Embed: NewEmbed().
				SetTitle("Remind").
				SetDescription("You asked me to remind you this "+humanize.Time(added)+"!").
				AddField("Reminder", message).
				SetColor(0x1C1C1C).MessageEmbed,
		})

		for i := len(remindEntries) - 1; i >= 0; i-- {
			if remindEntries[i].UserID == userID && remindEntries[i].Message == message {
				remindEntries = append(remindEntries[:i], remindEntries[i+1:]...)
				break
			}
		}
	})
}
