package main

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/olebedev/when"
)

// RemindEntry stores information about a remind entry
type RemindEntry struct {
	UserID    string    `json:"userID"`
	ChannelID string    `json:"channelID"`
	Message   string    `json:"message"`
	Added     time.Time `json:"timeAdded"`
	When      time.Time `json:"timeRemind"`
}

func commandRemind(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	w := when.EN
	text := strings.Join(args, " ")
	now := time.Now()

	r, err := w.Parse(text, now)
	if err != nil || r == nil {
		return NewErrorEmbed("Remind Error", "There was an error figuring out what time to remind you with this message at.")
	}

	remindWhen(env.User.ID, env.Channel.ID, text, now, r.Time, now)

	return NewGenericEmbed("Remind", humanize.Time(r.Time)+", we will remind you with the following message:\n\n```"+r.Source+"```")
}

func remindWhen(userID, channelID, message string, added, when, now time.Time) {
	remindEntries = append(remindEntries, RemindEntry{UserID: userID, ChannelID: channelID, Message: message, Added: added, When: when})

	waitDuration := when.Sub(now)
	time.AfterFunc(waitDuration, func() {
		botData.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content: "<@!" + userID + "> :alarm_clock:",
			Embed: NewEmbed().
				SetTitle("Remind").
				SetDescription("You asked me to remind you this "+humanize.Time(added)+"!").
				AddField("Message", message).
				SetColor(0x1C1C1C).MessageEmbed,
		})

		for i := range remindEntries {
			if remindEntries[i].UserID == userID && remindEntries[i].Message == message {
				remindEntries[i] = remindEntries[len(remindEntries)-1]
				remindEntries = remindEntries[:len(remindEntries)-1]
			}
		}
	})
}
