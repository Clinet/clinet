package main

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func commandRoll(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	random := rand.Intn(6) + 1
	return NewGenericEmbed("Roll", "You rolled a "+strconv.Itoa(random)+"!")
}
func commandDoubleRoll(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	random1 := rand.Intn(6) + 1
	random2 := rand.Intn(6) + 1
	randomTotal := random1 + random2
	return NewGenericEmbed("Double Roll", "You rolled a "+strconv.Itoa(random1)+" and a "+strconv.Itoa(random2)+". The total is "+strconv.Itoa(randomTotal)+"!")
}
func commandCoinFlip(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	random := rand.Intn(2)
	switch random {
	case 0:
		return NewGenericEmbed("Coin Flip", "The coin landed on heads!")
	case 1:
		return NewGenericEmbed("Coin Flip", "The coin landed on tails!")
	}
	return nil
}

func commandHewwo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	message := strings.Join(args, " ")

	message = strings.Replace(message, "L", "W", -1)
	message = strings.Replace(message, "l", "w", -1)
	message = strings.Replace(message, "R", "W", -1)
	message = strings.Replace(message, "r", "w", -1)

	return NewGenericEmbed("Hewwo", message)
}
