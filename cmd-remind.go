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
	GuildID   string    `json:"guildID"`
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

	defer remindWhen(env.User.ID, env.Guild.ID, env.Channel.ID, text, now, r.Time, now)

	return NewEmbed().
		SetTitle("Remind").
		SetDescription("I will give you this reminder "+humanize.Time(r.Time)).
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
