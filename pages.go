package main

import (
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"
)

// Page returns an Embed of a page based on the page items given and the page number requested
func page(pageItems []*discordgo.MessageEmbedField, page, maxResults int) (*Embed, int, error) {
	if len(pageItems) == 0 {
		return nil, 0, fmt.Errorf("No page items found.")
	}
	if page <= 0 {
		return nil, 0, fmt.Errorf("Page number %d too low.", page)
	}
	if maxResults <= 0 {
		return nil, 0, fmt.Errorf("Maximum results %d too low.", maxResults)
	}
	if maxResults >= EmbedLimitField {
		return nil, 0, fmt.Errorf("Maximum results %d too high.", maxResults)
	}

	totalPages := int(math.Ceil(float64(len(pageItems)) / float64(maxResults)))
	if page > totalPages {
		return nil, totalPages, fmt.Errorf("Page number %d too high.", page)
	}

	low := (page - 1) * maxResults
	high := page * maxResults
	if high > len(pageItems) {
		high = len(pageItems)
	}

	pageItems = pageItems[low:high]
	pageEmbed := NewEmbed()
	pageEmbed.Fields = pageItems
	return pageEmbed, totalPages, nil
}
