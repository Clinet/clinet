package main

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	message := strings.Join(args, " ")

	var buf bytes.Buffer

	z := zalgo.NewCorrupter(&buf)
	z.Zalgo = func(n int, r rune, z *zalgo.Corrupter) bool {
		z.Up += 0.01
		z.Middle += complex(0.01, 0.01)
		z.Down += complex(real(z.Down)*0.1, 0)
		return false
	}
	z.Up = complex(0, 0.1)
	z.Middle = complex(0, 0.1)
	z.Down = complex(0.001, 0.3)

	fmt.Fprintln(z, message)
	zalgo := buf.String()

	return NewGenericEmbed("Zalgo", zalgo)
}

func commandScreenshot(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if env.UpdatedMessageEvent {
		return nil
	}

	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	siteURL, err := url.Parse(args[0])
	if err != nil {
		return NewErrorEmbed("Screenshot Error", "Unknown address format ``"+args[0]+"``.")
	}
	if siteURL.Scheme == "" {
		siteURL.Scheme = "http" //By standard, SSL-enabled sites should automatically redirect to https if needed
	}
	website := siteURL.String()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://image.thum.io/get/maxAge/0/width/2000/noanimate/fullpage/%s", website), nil)
	if err != nil {
		return NewErrorEmbed("Screenshot Error", "The website ``"+args[0]+"`` does not exist or is currently unreachable.")
	}
	req.Header.Set("User-Agent", "Clinet/"+GitCommitFull)

	resp, err := client.Do(req)
	if err != nil {
		return NewErrorEmbed("Screenshot Error", "The website ``"+args[0]+"`` does not exist or is currently unreachable.")
	}

	var screenshotImage image.Image

	switch resp.Header.Get("content-type") {
	case "image/gif":
		gifAnim, err := gif.DecodeAll(resp.Body)
		if err != nil {
			return NewErrorEmbed("Screenshot Error", "The API failed to respond with a valid screenshot.")
		}
		screenshotImage = gifAnim.Image[len(gifAnim.Image)-1]
	case "image/png", "image/jpeg":
		srcImage, _, err := image.Decode(resp.Body)
		if err != nil {
			return NewErrorEmbed("Screenshot Error", "The API failed to respond with a valid screenshot.")
		}
		screenshotImage = srcImage
	default:
		return NewErrorEmbed("Screenshot Error", "The API failed to respond in an expected way.")
	}

	var outImage bytes.Buffer
	err = png.Encode(&outImage, screenshotImage)
	if err != nil {
		return NewErrorEmbed("Screenshot Error", "Unexpected error processing screenshot.")
	}

	imageName := website
	imageName = strings.Replace(imageName, "://", "_", -1)
	imageName = strings.Replace(imageName, "/", "_", -1)
	imageName += fmt.Sprintf("-%d", time.Now().Unix())
	imageName = "clinet-screenshot_" + imageName + ".png"
	_, err = botData.DiscordSession.ChannelMessageSendComplex(env.Channel.ID, &discordgo.MessageSend{
		File: &discordgo.File{
			Name:   imageName,
			Reader: &outImage,
		},
		Embed: &discordgo.MessageEmbed{
			Title:       "Screenshot",
			Description: website,
			Image: &discordgo.MessageEmbedImage{
				URL: "attachment://" + imageName,
			},
		},
	})
	if err != nil {
		return NewErrorEmbed("Screenshot Error", "Unexpected error uploading screenshot.")
	}
	return nil
}
