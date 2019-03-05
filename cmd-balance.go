package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
)

func commandBalance(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(env.Message.Mentions) > 0 {
		var mentions []*discordgo.User
		for _, mention := range env.Message.Mentions {
			unique := true
			for _, uniqueMention := range mentions {
				if uniqueMention.ID == mention.ID {
					unique = false
					break
				}
			}
			if unique {
				mentions = append(mentions, mention)
			}
		}
		var balances []string
		for _, mention := range mentions {
			if _, exists := userSettings[mention.ID]; !exists {
				balances = append(balances, "<@!"+mention.ID+">: $0")
			} else {
				balances = append(balances, "<@!"+mention.ID+">: $"+strconv.Itoa(userSettings[mention.ID].Balance))
			}
		}
		return NewGenericEmbedAdvanced("Balance", "The balances of the mentioned users are available below:\n\n"+strings.Join(balances, "\n"), 0x85BB65)
	}

	if userSettings[env.User.ID].DailyNext.IsZero() {
		return NewGenericEmbedAdvanced("Balance", "Your current balance is __$"+strconv.Itoa(userSettings[env.User.ID].Balance)+"__!\n\nYou may run "+botData.CommandPrefix+"daily to receive your first __$200__ daily credits.", 0x85BB65)
	}

	return NewGenericEmbedAdvanced("Balance", "Your current balance is __$"+strconv.Itoa(userSettings[env.User.ID].Balance)+"__!\n\nYou may receive your next __$200__ daily credits approximately "+humanize.Time(userSettings[env.User.ID].DailyNext)+".", 0x85BB65)
}

func commandDaily(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if userSettings[env.User.ID].DailyNext.IsZero() {
		userSettings[env.User.ID].Balance += 5000
		userSettings[env.User.ID].DailyNext = time.Now().Add(time.Hour * 24)
		return NewGenericEmbedAdvanced("Daily", "You received your __$200__ daily credits!\n\nAs a bonus for your first daily, you received an additional __$4800__ credits!", 0x85BB65)
	}

	if time.Now().After(userSettings[env.User.ID].DailyNext) {
		userSettings[env.User.ID].Balance += 200
		userSettings[env.User.ID].DailyNext = time.Now().Add(time.Hour * 24)
		return NewGenericEmbedAdvanced("Daily", "You received your __$200__ daily credits!", 0x85BB65)
	}

	return NewGenericEmbedAdvanced("Daily", "You have already received your __$200__ daily credits!\n\nYou may receive your next __$200__ daily credits approximately "+humanize.Time(userSettings[env.User.ID].DailyNext)+".", 0x85BB65)
}

func commandTransfer(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(env.Message.Mentions) <= 0 {
		return NewErrorEmbed("Transfer Error", "You must specify a user to transfer credits to.")
	}
	if len(env.Message.Mentions) > 1 {
		return NewErrorEmbed("Transfer Error", "You cannot specify more than one user to transfer credits to.")
	}

	credits, err := strconv.Atoi(args[0])
	if err != nil {
		return NewErrorEmbed("Transfer Error", "``"+args[0]+"`` is not a valid number.")
	}

	if credits <= 0 {
		return NewErrorEmbed("Transfer Error", "You cannot transfer less than __$1__ in credits.")
	}

	if credits > userSettings[env.User.ID].Balance {
		return NewErrorEmbed("Transfer Error", "You have insufficient credits to perform this transfer.")
	}

	target := env.Message.Mentions[0]
	if target.Bot {
		return NewErrorEmbed("Transfer Error", "You cannot transfer credits to a bot!")
	}
	initializeUserSettings(target.ID)

	userSettings[env.User.ID].Balance -= credits
	userSettings[target.ID].Balance += credits

	return NewGenericEmbed("Transfer", "Successfully transferred __$"+args[0]+"__ in credits to <@!"+target.ID+">!")
}
