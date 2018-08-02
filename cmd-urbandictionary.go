package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"

	"github.com/JoshuaDoes/urbandictionary"
	//prettyTime "github.com/andanhm/go-prettytime"
	"github.com/bwmarrin/discordgo"
)

func commandUrbanDictionary(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	results, err := urbandictionary.Query(args[0])
	if err != nil {
		return NewErrorEmbed("Urban Dictionary Error", "There was an error getting a result for that term.")
	}

	linkExp := regexp.MustCompile(`\[([^\]]*)\]`)
	linkExpFunc := func(s string) string {
		ss := linkExp.FindStringSubmatch(s)
		if len(ss) == 0 {
			return s
		}
		hyperlink := "https://www.urbandictionary.com/define.php?term=" + url.QueryEscape(ss[1])
		return fmt.Sprintf("%s(%s)", s, hyperlink)
	}

	result := results.Results[0]
	resultEmbed := NewEmbed().
		SetTitle("Urban Dictionary - "+result.Word).
		SetDescription(linkExp.ReplaceAllStringFunc(result.Definition, linkExpFunc)).
		AddField("Example", linkExp.ReplaceAllStringFunc(result.Example, linkExpFunc)).
		AddField("Author", result.Author).
		AddField("Stats", "\U0001f44d "+strconv.Itoa(result.ThumbsUp)+" \U0001f44e "+strconv.Itoa(result.ThumbsDown)).
		SetFooter("Results from Urban Dictionary.", "https://res.cloudinary.com/hrscywv4p/image/upload/c_limit,fl_lossy,h_300,w_300,f_auto,q_auto/v1/1194347/vo5ge6mdw4creyrgaq2m.png").
		SetURL(result.Permalink).
		SetColor(0xDC2A26).MessageEmbed

	//Oddly enough, this isn't wanting to return anything for the moment so it's staying commented
	//	date := prettyTime.Format(result.Date)
	//	if date != "" {
	//		resultEmbed.Fields = append(resultEmbed.Fields, &discordgo.MessageEmbedField{Name: "Date Written", Value: date})
	//	}

	return resultEmbed
}