package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kortschak/zalgo"
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

func commandZalgo(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	z := zalgo.NewCorrupter(writer)
	z.Zalgo = func(n int, r rune, z *zalgo.Corrupter) bool {
		z.Up += 0.01
		z.Middle += complex(0.01, 0.01)
		z.Down += complex(real(z.Down)*0.1, 0)
		return false
	}
	z.Up = complex(0, 0.2)
	z.Middle = complex(0, 0.2)
	z.Down = complex(0.001, 0.3)

	fmt.Fprint(writer, []byte(strings.Join(args, " ")))
	zalgo := buf.String()

	return NewGenericEmbed("Zalgo", string(zalgo))
}
