package main

import (
	"fmt"
	"math"

	"github.com/bwmarrin/discordgo"
)

type PagedList struct {
	Items      []*discordgo.MessageEmbedField `json:"pageItems"`  //The items to be paginated
	MaxResults int                            `json:"maxResults"` //The maximum results per page
	PageNumber int                            `json:"pageNumber"` //The current page number
	TotalPages int                            `json:"totalPages"` //The total amount of pages available
	FirstPage  bool                           `json:"firstPage"`  //Whether or not we're on the first page
	LastPage   bool                           `json:"lastPage"`   //Whether or not we're on the last page
	Ext        *map[string]interface{}        `json:"ext"`        //Extra data for the current paged list, used to keep track of things for whatever is implementing the paged list
}

// NewPageList returns a new paginated items list
func NewPagedList(items []*discordgo.MessageEmbedField, maxResults int) (*PagedList, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("No page items found.")
	}

	return &PagedList{
		Items:      items,
		MaxResults: maxResults,
		PageNumber: 1,
		TotalPages: int(math.Ceil(float64(len(items)) / float64(maxResults))),
		FirstPage:  true,
		LastPage:   false,
		Ext:        new(map[string]interface{}),
	}, nil
}

// GetPage returns a certain page from a paginated items list
func (pagedList *PagedList) GetPage(pageNumber int) (*Embed, error) {
	if err := pagedList.Check(); err != nil {
		return nil, err
	}
	if pageNumber <= 0 {
		return nil, fmt.Errorf("Page number %d too low.", pageNumber)
	}
	if pageNumber > int(math.Ceil(float64(len(pagedList.Items))/float64(pagedList.MaxResults))) {
		return nil, fmt.Errorf("Page number %d too high.", pagedList.PageNumber)
	}

	pagedList.TotalPages = int(math.Ceil(float64(len(pagedList.Items)) / float64(pagedList.MaxResults)))
	if pageNumber > pagedList.TotalPages {
		return nil, fmt.Errorf("Page number %d too high.", pageNumber)
	}

	pagedList.PageNumber = pageNumber
	if pageNumber == 1 {
		pagedList.FirstPage = true
		pagedList.LastPage = false
	} else if pageNumber == pagedList.TotalPages {
		pagedList.FirstPage = false
		pagedList.LastPage = true
	} else {
		pagedList.FirstPage = false
		pagedList.LastPage = false
	}

	low := (pagedList.PageNumber - 1) * pagedList.MaxResults
	high := pagedList.PageNumber * pagedList.MaxResults
	if high > len(pagedList.Items) {
		high = len(pagedList.Items)
	}

	pageItems := pagedList.Items[low:high]
	pageEmbed := NewEmbed()
	pageEmbed.Fields = pageItems
	return pageEmbed, nil
}

// GetCurrentPage returns the current page
func (pagedList *PagedList) GetCurrentPage() (*Embed, error) {
	return pagedList.GetPage(pagedList.PageNumber)
}

// GetNextPage returns the next page
func (pagedList *PagedList) GetNextPage() (*Embed, error) {
	return pagedList.GetPage(pagedList.PageNumber + 1)
}

// GetPreviousPage returns the next page
func (pagedList *PagedList) GetPreviousPage() (*Embed, error) {
	return pagedList.GetPage(pagedList.PageNumber - 1)
}

func (pagedList *PagedList) Check() error {
	if len(pagedList.Items) == 0 {
		return fmt.Errorf("No page items found.")
	}
	if pagedList.MaxResults <= 0 {
		return fmt.Errorf("Maximum results %d too low.", pagedList.MaxResults)
	}
	if pagedList.MaxResults >= EmbedLimitField {
		return fmt.Errorf("Maximum results %d too high.", pagedList.MaxResults)
	}
	if pagedList.PageNumber <= 0 {
		return fmt.Errorf("Page number %d too low.", pagedList.PageNumber)
	}
	if pagedList.PageNumber > int(math.Ceil(float64(len(pagedList.Items))/float64(pagedList.MaxResults))) {
		return fmt.Errorf("Page number %d too high.", pagedList.PageNumber)
	}
	return nil
}

// page returns an Embed of a page based on the page items given and the page number requested
// Kept for backwards compatibility with the previous pagination implementation, deprecated until removal
func page(pageItems []*discordgo.MessageEmbedField, page, maxResults int) (*Embed, int, error) {
	pagedList, err := NewPagedList(pageItems, maxResults)
	if err != nil {
		return nil, 0, err
	}
	currentPage, err := pagedList.GetPage(page)
	if err != nil {
		return nil, 0, err
	}
	return currentPage, pagedList.TotalPages, nil
}
