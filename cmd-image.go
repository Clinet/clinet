package main

import (
	"bytes"
	"image"
	"image/png"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/gift"
)

func commandImage(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	if len(env.Message.Attachments) > 0 {
		for _, attachment := range env.Message.Attachments {
			srcImageURL := attachment.URL
			srcImageHTTP, err := http.Get(srcImageURL)
			if err != nil {
				return NewErrorEmbed("Image Error", "Unable to fetch image.")
			}
			srcImage, _, err := image.Decode(srcImageHTTP.Body)
			if err != nil {
				return NewErrorEmbed("Image Error", "Unable to decode image.")
			}

			g := &gift.GIFT{}
			var outImage bytes.Buffer

			switch args[0] {
			case "fliphorizontal":
				g = gift.New(gift.FlipHorizontal())
			case "flipvertical":
				g = gift.New(gift.FlipVertical())
			case "grayscale", "greyscale":
				g = gift.New(gift.Grayscale())
			case "invert":
				g = gift.New(gift.Invert())
			case "rotate90":
				g = gift.New(gift.Rotate90())
			case "rotate180":
				g = gift.New(gift.Rotate180())
			case "rotate270":
				g = gift.New(gift.Rotate270())
			case "sobel":
				g = gift.New(gift.Sobel())
			case "transpose":
				g = gift.New(gift.Transpose())
			case "transverse":
				g = gift.New(gift.Transverse())
			}

			dstImage := image.NewRGBA(g.Bounds(srcImage.Bounds()))
			g.Draw(dstImage, srcImage)

			err = png.Encode(&outImage, dstImage)
			if err != nil {
				return NewErrorEmbed("Image Error", "Unable to encode processed image.")
			}
			_, err = botData.DiscordSession.ChannelMessageSendComplex(env.Channel.ID, &discordgo.MessageSend{
				File: &discordgo.File{
					Name:   args[0] + ".png",
					Reader: &outImage,
				},
				Embed: &discordgo.MessageEmbed{
					Title: "Processed Image",
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://" + args[0] + ".png",
					},
				},
			})
			if err != nil {
				return NewErrorEmbed("Image Error", "Unable to upload processed image.")
			}
		}
	}
	return NewErrorEmbed("Image Error", "You must upload an image to process.")
}
