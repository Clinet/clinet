package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/gift"
	"github.com/go-playground/colors"
)

func commandImageAdv(args []CommandArgument, env *CommandEnvironment) *discordgo.MessageEmbed {
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

			g := gift.New()
			var outImage bytes.Buffer

			backgroundColor := color.RGBA{0, 0, 0, 0}
			interpolation := gift.NearestNeighborInterpolation
			resampling := gift.NearestNeighborResampling

			width := srcImage.Bounds().Max.X
			height := srcImage.Bounds().Max.Y

			for _, effect := range args {
				switch effect.Name {
				case "bg", "bgcolor", "bgcolour", "backgroundcolor", "backgroundcolour":
					newBackgroundColor, err := colors.Parse(effect.Value)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid background color ``"+effect.Value+"``.")
					}
					newBackgroundColorRGBA := newBackgroundColor.ToRGBA()
					alpha := uint8(newBackgroundColorRGBA.A * 0xFF)

					backgroundColor = color.RGBA{R: newBackgroundColorRGBA.R, G: newBackgroundColorRGBA.G, B: newBackgroundColorRGBA.B, A: alpha}

					fmt.Println(fmt.Sprintf("%v", newBackgroundColor))
					fmt.Println(fmt.Sprintf("%v", newBackgroundColorRGBA))
					fmt.Println(fmt.Sprintf("%v", backgroundColor))
				case "brightness":
					brightness, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid brightness percentage ``"+effect.Value+"``.")
					}
					brightness -= 100
					g.Add(gift.Brightness(float32(brightness)))
				case "contrast":
					contrast, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid contrast percentage ``"+effect.Value+"``.")
					}
					contrast -= 100
					g.Add(gift.Contrast(float32(contrast)))
				case "f", "flip":
					switch effect.Value {
					case "h", "horizontal", "left", "right":
						g.Add(gift.FlipHorizontal())
					case "v", "vertical", "up", "down":
						g.Add(gift.FlipVertical())
					default:
						return NewErrorEmbed("Image Error", "Invalid flip direction ``"+effect.Value+"``.")
					}
				case "gamma":
					gamma, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid gamma percentage ``"+effect.Value+"``.")
					}
					gamma /= 100
					g.Add(gift.Gamma(float32(gamma)))
				case "gaussian", "gaussianblur":
					gaussian, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid gaussian blur percentage ``"+effect.Value+"``.")
					}
					gaussian /= 100
					g.Add(gift.GaussianBlur(float32(gaussian)))
				case "grayscale", "greyscale":
					g.Add(gift.Grayscale())
				case "height":
					newHeight, err := strconv.Atoi(effect.Value)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid height integer ``"+effect.Value+"``.")
					}
					height = newHeight
				case "interpolation":
					switch effect.Value {
					case "c", "cubic":
						interpolation = gift.CubicInterpolation
					case "l", "linear":
						interpolation = gift.LinearInterpolation
					case "nn", "nearestneighbor", "nearestneighbour", "nearest":
						interpolation = gift.NearestNeighborInterpolation
					default:
						return NewErrorEmbed("Image Error", "Invalid interpolation ``"+effect.Value+"``.")
					}
				case "invert":
					g.Add(gift.Invert())
				case "pixelate":
					pixelate, err := strconv.Atoi(effect.Value)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid pixelation integer ``"+effect.Value+"``.")
					}
					g.Add(gift.Pixelate(pixelate))
				case "resampling":
					switch effect.Value {
					case "b", "box":
						resampling = gift.BoxResampling
					case "c", "cubic":
						resampling = gift.CubicResampling
					case "lanczos":
						resampling = gift.LanczosResampling
					case "linear":
						resampling = gift.LinearResampling
					case "nn", "nearestneighbor", "nearestneighbour", "nearest":
						resampling = gift.NearestNeighborResampling
					default:
						return NewErrorEmbed("Image Error", "Invalid resampling ``"+effect.Value+"``.")
					}
				case "rotate":
					angle, err := strconv.ParseFloat(effect.Value, 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid rotation angle ``"+effect.Value+"``.")
					}
					g.Add(gift.Rotate(float32(angle), backgroundColor, interpolation))
				case "saturation":
					saturation, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid saturation percentage ``"+effect.Value+"``.")
					}
					saturation -= 100
					g.Add(gift.Saturation(float32(saturation)))
				case "sepia":
					sepia, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid sepia percentage ``"+effect.Value+"``.")
					}
					g.Add(gift.Sepia(float32(sepia)))
				case "sobel":
					g.Add(gift.Sobel())
				case "threshold":
					threshold, err := strconv.ParseFloat(strings.TrimSuffix(effect.Value, "%"), 32)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid threshold percentage ``"+effect.Value+"``.")
					}
					g.Add(gift.Threshold(float32(threshold)))
				case "transpose":
					g.Add(gift.Transpose())
				case "transverse":
					g.Add(gift.Transverse())
				case "width":
					newWidth, err := strconv.Atoi(effect.Value)
					if err != nil {
						return NewErrorEmbed("Image Error", "Invalid width integer ``"+effect.Value+"``.")
					}
					width = newWidth
				default:
					return NewErrorEmbed("Image Error", "Unknown effect ``"+effect.Name+"``.")
				}
			}

			g.Add(gift.Resize(width, height, resampling))

			dstImage := image.NewRGBA(g.Bounds(srcImage.Bounds()))
			g.Draw(dstImage, srcImage)

			err = png.Encode(&outImage, dstImage)
			if err != nil {
				return NewErrorEmbed("Image Error", "Unable to encode processed image.")
			}
			_, err = botData.DiscordSession.ChannelMessageSendComplex(env.Channel.ID, &discordgo.MessageSend{
				File: &discordgo.File{
					Name:   "clinet-processed.png",
					Reader: &outImage,
				},
				Embed: &discordgo.MessageEmbed{
					Title: "Processed Image",
					Image: &discordgo.MessageEmbedImage{
						URL: "attachment://clinet-processed.png",
					},
				},
			})
			if err != nil {
				return NewErrorEmbed("Image Error", "Unable to upload processed image.")
			}
			return nil
		}
	}
	return NewErrorEmbed("Image Error", "You must upload an image to process.")
}
