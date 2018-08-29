package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var (
	regexCmdPlay     string = "(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)" //@Clinet Play Raining Tacos
	regexCmdPlayComp *regexp.Regexp
)

func queryWolframAlpha(query string) (*discordgo.MessageEmbed, error) {
	debugLog("[Wolfram|Alpha] Getting result for query ["+query+"]...", false)
	queryResultData, err := botData.BotClients.Wolfram.GetQueryResult(query, nil)
	if err != nil {
		debugLog("[Wolfram|Alpha] Error getting query result: "+fmt.Sprintf("%v", err), false)
		return nil, errors.New("Error getting response from Wolfram|Alpha.")
	}

	result := queryResultData.QueryResult
	pods := result.Pods
	if len(pods) == 0 {
		debugLog("[Wolfram|Alpha] Error getting pods from query", false)
		return nil, errors.New("Error getting pods from query.")
	}

	fields := []*discordgo.MessageEmbedField{}

	for _, pod := range pods {
		podTitle := pod.Title
		if wolframIsPodDenied(podTitle) {
			debugLog("[Wolfram|Alpha] Denied pod: "+podTitle, false)
			continue
		}

		subPods := pod.SubPods
		if len(subPods) > 0 { //Skip this pod if no subpods are found
			for _, subPod := range subPods {
				plaintext := subPod.Plaintext
				if plaintext != "" {
					fields = append(fields, &discordgo.MessageEmbedField{Name: podTitle, Value: plaintext, Inline: true})
				}
			}
		}
	}

	if len(fields) == 0 { //No results were found
		debugLog("[Wolfram|Alpha] Error getting legal data from Wolfram|Alpha", false)
		return nil, errors.New("Error getting legal data from Wolfram|Alpha.")
	} else {
		wolframEmbed := NewEmbed().
			SetColor(0xDA0E1A).
			SetFooter("Results from Wolfram|Alpha.", "https://upload.wikimedia.org/wikipedia/en/thumb/8/83/Wolfram_Alpha_December_2016.svg/257px-Wolfram_Alpha_December_2016.svg.png").MessageEmbed
		wolframEmbed.Fields = fields
		return wolframEmbed, nil
	}
}
func wolframIsPodDenied(podTitle string) bool {
	for _, deniedPodTitle := range botData.BotOptions.WolframDeniedPods {
		if deniedPodTitle == podTitle {
			return true //Pod is denied
		}
	}
	return false //Pod is not denied
}

func queryDuckDuckGo(query string) (*discordgo.MessageEmbed, error) {
	debugLog("[DuckDuckGo] Getting result for query ["+query+"]...", false)
	queryResult, err := botData.BotClients.DuckDuckGo.GetQueryResult(query)
	if err != nil {
		debugLog("[DuckDuckGo] Error getting query result: "+fmt.Sprintf("%v", err), false)
		return nil, errors.New("Error getting response from DuckDuckGo.")
	}

	result := ""
	if queryResult.Definition != "" {
		result = queryResult.Definition
	} else if queryResult.Answer != "" {
		result = queryResult.Answer
	} else if queryResult.AbstractText != "" {
		result = queryResult.AbstractText
	}
	if result == "" {
		debugLog("[DuckDuckGo] Error getting allowed result from response", false)
		return nil, errors.New("Error getting allowed result from response.")
	}

	duckduckgoEmbed := NewEmbed().
		SetTitle(queryResult.Heading).
		SetDescription(result).
		SetColor(0xDF5730).
		SetFooter("Results from DuckDuckGo.", "https://upload.wikimedia.org/wikipedia/en/9/90/The_DuckDuckGo_Duck.png").MessageEmbed
	if queryResult.Image != "" {
		duckduckgoEmbed.Image = &discordgo.MessageEmbedImage{URL: queryResult.Image}
	}
	return duckduckgoEmbed, nil
}

func queryImgur(url string) (*discordgo.MessageEmbed, error) {
	imgurInfo, _, err := botData.BotClients.Imgur.GetInfoFromURL(url)
	if err != nil {
		debugLog("[Imgur] Error getting info from URL ["+url+"]", false)
		return nil, errors.New("Error getting info from URL.")
	}
	if imgurInfo.Image != nil {
		debugLog("[Imgur] Detected image from URL ["+url+"]", false)
		imgurImage := imgurInfo.Image
		imgurEmbed := NewEmbed().
			SetTitle(imgurImage.Title).
			SetDescription(imgurImage.Description).
			AddField("Views", strconv.Itoa(imgurImage.Views)).
			AddField("NSFW", strconv.FormatBool(imgurImage.Nsfw)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.Album != nil {
		debugLog("[Imgur] Detected album from URL ["+url+"]", false)
		imgurAlbum := imgurInfo.Album
		imgurEmbed := NewEmbed().
			SetTitle(imgurAlbum.Title).
			SetDescription(imgurAlbum.Description).
			AddField("Uploader", imgurAlbum.AccountURL).
			AddField("Image Count", strconv.Itoa(imgurAlbum.ImagesCount)).
			AddField("Views", strconv.Itoa(imgurAlbum.Views)).
			AddField("NSFW", strconv.FormatBool(imgurAlbum.Nsfw)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.GImage != nil {
		debugLog("[Imgur] Detected gallery image from URL ["+url+"]", false)
		imgurGImage := imgurInfo.GImage
		imgurEmbed := NewEmbed().
			SetTitle(imgurGImage.Title).
			SetDescription(imgurGImage.Description).
			AddField("Topic", imgurGImage.Topic).
			AddField("Uploader", imgurGImage.AccountURL).
			AddField("Views", strconv.Itoa(imgurGImage.Views)).
			AddField("NSFW", strconv.FormatBool(imgurGImage.Nsfw)).
			AddField("Comment Count", strconv.Itoa(imgurGImage.CommentCount)).
			AddField("Upvotes", strconv.Itoa(imgurGImage.Ups)).
			AddField("Downvotes", strconv.Itoa(imgurGImage.Downs)).
			AddField("Points", strconv.Itoa(imgurGImage.Points)).
			AddField("Score", strconv.Itoa(imgurGImage.Score)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else if imgurInfo.GAlbum != nil {
		debugLog("[Imgur] Detected gallery album from URL ["+url+"]", false)
		imgurGAlbum := imgurInfo.GAlbum
		imgurEmbed := NewEmbed().
			SetTitle(imgurGAlbum.Title).
			SetDescription(imgurGAlbum.Description).
			AddField("Topic", imgurGAlbum.Topic).
			AddField("Uploader", imgurGAlbum.AccountURL).
			AddField("Views", strconv.Itoa(imgurGAlbum.Views)).
			AddField("NSFW", strconv.FormatBool(imgurGAlbum.Nsfw)).
			AddField("Comment Count", strconv.Itoa(imgurGAlbum.CommentCount)).
			AddField("Upvotes", strconv.Itoa(imgurGAlbum.Ups)).
			AddField("Downvotes", strconv.Itoa(imgurGAlbum.Downs)).
			AddField("Points", strconv.Itoa(imgurGAlbum.Points)).
			AddField("Score", strconv.Itoa(imgurGAlbum.Score)).
			SetColor(0x89C623).MessageEmbed
		return imgurEmbed, nil
	} else {
		debugLog("[Imgur] Error detecting Imgur type from URL ["+url+"]", false)
		return nil, errors.New("Error detecting Imgur URL type.")
	}
	return nil, errors.New("Error detecting Imgur URL type.")
}
