package main

/*
* This voice handler is heavily based on https://github.com/iheanyi/bandcamp-go and
* should be credited for the work put into it despite its MIT licensing.
 */

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

// Bandcamp exports the methods required to access the Bandcamp service
type Bandcamp struct {
}

// GetName returns the service's name
func (*Bandcamp) GetName() string {
	return "Bandcamp"
}

// GetColor returns the service's color
func (*Bandcamp) GetColor() int {
	return 0x629AA9
}

// TestURL tests if the given URL is a Bandcamp album or track URL
func (*Bandcamp) TestURL(url string) (bool, error) {
	test, err := regexp.MatchString("^https://[a-z0-9\\\\-]+?\\.bandcamp\\.com/(track|album)/[a-z0-9\\\\-]+?/?$", url)
	return test, err
}

// GetMetadata returns the metadata for a given Bandcamp track URL
func (*Bandcamp) GetMetadata(url string) (*Metadata, error) {
	album, err := bandcampGetAlbum(url)
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		Title:        album.Tracks[0].Title,
		DisplayURL:   "https://" + album.Artist + ".bandcamp.com" + album.Tracks[0].TitleLink,
		StreamURL:    album.Tracks[0].Files.MP3128,
		Duration:     album.Tracks[0].Duration,
		ArtworkURL:   "https://f4.bcbits.com/img/a0" + strconv.Itoa(album.ArtID) + "_10.jpg",
		ThumbnailURL: "https://f4.bcbits.com/img/a0" + strconv.Itoa(album.ArtID) + "_10.jpg",
	}

	trackArtist := &MetadataArtist{
		Name: album.Artist,
		URL:  "https://" + album.Artist + ".bandcamp.com",
	}
	metadata.Artists = append(metadata.Artists, *trackArtist)

	return metadata, nil
}

type bandcampAlbum struct {
	Artist string `json:"artist"`
	ArtID  int    `json:"art_id"`
	Info   struct {
		Title string `json:"title"`
	} `json:"current"`
	Tracks []*bandcampTrack `json:"trackinfo"`
	URL    string           `json:"url"`
}

type bandcampTrack struct {
	Duration float64 `json:"duration"`
	Files    struct {
		MP3128 string `json:"mp3-128"`
	} `json:"file"`
	Title     string `json:"title"`
	TitleLink string `json:"title_link"`
}

func bandcampGetAlbum(url string) (*bandcampAlbum, error) {
	_, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	var albumInfo map[string]interface{}

	doc.Find(".yui-skin-sam script").Each(func(i int, s *goquery.Selection) {
		if i == 1 {
			nodeText := s.Text()
			albumDataDef := strings.Split(nodeText, "var TralbumData = ")[1]
			albumData := strings.Split(albumDataDef, ";")[0]
			albumInfo, err = generateAlbumMap(albumData)
		}
	})
	if err != nil {
		return nil, err
	}

	var album *bandcampAlbum

	albumInfoJSON, err := json.MarshalIndent(albumInfo, "", "\t")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(albumInfoJSON, &album)
	if err != nil {
		return nil, err
	}

	if album == nil {
		return nil, fmt.Errorf("error finding album")
	}

	if len(album.Tracks) == 0 {
		return nil, fmt.Errorf("error finding tracks")
	}

	return album, nil
}

func generateAlbumMap(jsCode string) (map[string]interface{}, error) {
	fullCodeBlock := "albumData = " + jsCode

	vm := otto.New()
	vm.Run(fullCodeBlock)
	vm.Run(`
albumDataStr = JSON.stringify(albumData);
`)

	var albumMap map[string]interface{}

	value, err := vm.Get("albumDataStr")
	if err != nil {
		return albumMap, err
	}

	valueStr, err := value.ToString()
	if err != nil {
		return albumMap, err
	}

	jsonByteArray := []byte(valueStr)
	err = json.Unmarshal(jsonByteArray, &albumMap)

	if err != nil {
		return albumMap, err
	}

	return albumMap, nil
}
