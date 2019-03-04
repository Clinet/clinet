package main

import (
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	humanize "github.com/dustin/go-humanize"
)

func commandBalance(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if userSettings[env.User.ID].DailyNext.IsZero() {
		return NewGenericEmbedAdvanced("Balance", "Your current balance is __$"+strconv.Itoa(userSettings[env.User.ID].Balance)+"__!\n\nYou may run "+botData.CommandPrefix+"daily to receive your first __$200__ daily credits.", 0x85BB65)
	}
	return NewGenericEmbedAdvanced("Balance", "Your current balance is __$"+strconv.Itoa(userSettings[env.User.ID].Balance)+"__!\n\nYou may receive your next __$200__ daily credits "+humanize.Time(userSettings[env.User.ID].DailyNext)+".", 0x85BB65)
}

func commandDaily(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if userSettings[env.User.ID].DailyNext.IsZero() {
		userSettings[env.User.ID].Balance += 5000
		userSettings[env.User.ID].DailyNext = time.Now().Add(time.Hour * 24)
		return NewGenericEmbedAdvanced("Daily", "You received your __$200__ daily credits!\n\nAs a bonus for your first daily, you received an additional __$4800__ credits!", 0x85BB65)
	}

	elapsed := time.Now().Sub(userSettings[env.User.ID].DailyNext)

	if elapsed.Hours() < 24 {
		return NewGenericEmbedAdvanced("Daily", "You have already received your __$200__ daily credits!\n\nYou may receive your next __$200__ daily credits "+humanize.Time(userSettings[env.User.ID].DailyNext)+".", 0x85BB65)
	}

	userSettings[env.User.ID].Balance += 200
	userSettings[env.User.ID].DailyNext = time.Now().Add(time.Hour * 24)
	return NewGenericEmbedAdvanced("Daily", "You received your __$200__ daily credits!", 0x85BB65)
}
