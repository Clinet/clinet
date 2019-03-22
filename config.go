package main

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/JoshuaDoes/duckduckgolang"
	"github.com/JoshuaDoes/go-soundcloud"
	"github.com/JoshuaDoes/go-wolfram"
	"github.com/JoshuaDoes/spotigo"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
	"github.com/jonas747/dca"
	"github.com/koffeinsource/go-imgur"
	"github.com/nishanths/go-xkcd"
	"github.com/rhnvrm/lyric-api-go"
	"github.com/superwhiskers/fennel"
	"google.golang.org/api/youtube/v3"
)

// BotData stores all data for the bot
type BotData struct {
	BotClients           BotClients
	BotKeys              BotKeys               `json:"botKeys"`
	BotName              string                `json:"-"`
	BotOwnerID           string                `json:"botOwnerID"`
	BotOptions           BotOptions            `json:"botOptions"`
	BotToken             string                `json:"botToken"`
	BotInviteURL         string                `json:"botInviteURL"`
	BotDiscordURL        string                `json:"botDiscordURL"`
	BotDonationURL       string                `json:"botDonationURL"`
	BotSourceURL         string                `json:"botSourceURL"`
	CommandPrefix        string                `json:"cmdPrefix"`
	CustomResponses      []CustomResponseQuery `json:"customResponses"`
	CustomStatuses       []CustomStatus        `json:"customStatuses"`
	TipMessages          []TipMessage          `json:"tipMessages"`
	DebugMode            bool                  `json:"debugMode"`
	SendOwnerStackTraces bool                  `json:"sendOwnerStackTraces"`

	DiscordSession *discordgo.Session
	Commands       map[string]*Command
	NLPCommands    []*CommandNLP
	VoiceServices  []VoiceService
	QueryServices  []QueryService
	LastTipMessage int

	Updating bool
}

// BotClients stores available clients for the bot
type BotClients struct {
	DuckDuckGo *duckduckgo.Client
	GitHub     *github.Client
	Imgur      imgur.Client
	Lyrics     lyrics.Lyric
	SoundCloud *soundcloud.Client
	Spotify    *spotigo.Client
	Wolfram    *wolfram.Client
	XKCD       *xkcd.Client
	YouTube    *youtube.Service
	Ninty      *fennel.AccountServerClient
}

// BotKeys stores all bot keys for using external services
type BotKeys struct {
	DuckDuckGoAppName    string                   `json:"ddgAppName"`
	GeniusAccessToken    string                   `json:"geniusAccessToken"`
	ImgurClientID        string                   `json:"imgurClientID"`
	SoundCloudAppVersion string                   `json:"soundcloudAppVersion"`
	SoundCloudClientID   string                   `json:"soundcloudClientID"`
	SpotifyHost          string                   `json:"spotifyHost"`
	SpotifyPass          string                   `json:"spotifyPass"`
	WolframAppID         string                   `json:"wolframAppID"`
	YouTubeAPIKey        string                   `json:"youtubeAPIKey"`
	Ninty                fennel.ClientInformation `json:"ninty"`
}

// BotOptions stores all bot options
type BotOptions struct {
	MaxPingCount       int                `json:"maxPingCount"` //How many pings to test to determine the average ping
	HelpMaxResults     int                `json:"helpMaxResults"`
	SendTypingEvent    bool               `json:"sendTypingEvent"`
	UseCustomResponses bool               `json:"useCustomResponses"`
	UseDuckDuckGo      bool               `json:"useDuckDuckGo"`
	UseGitHub          bool               `json:"useGitHub"`
	UseImgur           bool               `json:"useImgur"`
	UseLyrics          bool               `json:"useLyrics"`
	UseNinty           bool               `json:"useNinty"`
	UseSoundCloud      bool               `json:"useSoundCloud"`
	UseSpotify         bool               `json:"useSpotify"`
	UseWolframAlpha    bool               `json:"useWolframAlpha"`
	UseXKCD            bool               `json:"useXKCD"`
	UseYouTube         bool               `json:"useYouTube"`
	WolframDeniedPods  []string           `json:"wolframDeniedPods"`
	YouTubeMaxResults  int                `json:"youtubeMaxResults"`
	SpotifyMaxResults  int                `json:"spotifyMaxResults"`
	AudioEncoding      *dca.EncodeOptions `json:"audioEncoding"`
	API                APIConfig          `json:"api"`
}

// API stores configurations for the API
type APIConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
}

// CustomResponseQuery stores a custom response
type CustomResponseQuery struct {
	Expression   string `json:"expression"`
	Regexp       *regexp.Regexp
	Responses    []CustomResponseReply    `json:"responses"`
	CmdResponses []CustomResponseReplyCmd `json:"cmdResponses"`
}

// CustomResponseReply stores a custom response's reply
type CustomResponseReply struct {
	ResponseEmbed *discordgo.MessageEmbed `json:"responseEmbed"`
}

// CustomResponseReplyCmd stores a custom response's command to execute
type CustomResponseReplyCmd struct {
	CommandName string   `json:"commandName"`
	Arguments   []string `json:"args"`
}

// CustomStatus stores a custom status for the bot's presence status
type CustomStatus struct {
	Type   int    `json:"type"`
	Status string `json:"status"`
	URL    string `json:"url,omitempty"`
}

// TipMessage stores a tip message for how to use a specific feature in the bot
type TipMessage struct {
	FeatureName string   `json:"featureName"`
	DidYouKnow  string   `json:"didYouKnow"`
	HowTo       string   `json:"howTo"`
	Examples    []string `json:"examples"`
}

// PrepConfig checks the configuration for consistency and invalid errors
func (configData *BotData) PrepConfig() error {
	//Bot config checks
	if configData.BotToken == "" {
		return errors.New("config:{botName: \"\"}")
	}
	if configData.CommandPrefix == "" {
		return errors.New("config:{cmdPrefix: \"\"}")
	}

	//Value checks
	if configData.BotOptions.MaxPingCount > 5 || configData.BotOptions.MaxPingCount <= 0 {
		return errors.New("config:{botOptions:{maxPingCount}} must be 1, 2, 3, 4, or 5")
	}
	if configData.BotOptions.HelpMaxResults > EmbedLimitField || configData.BotOptions.HelpMaxResults <= 0 {
		return errors.New("config:{botOptions:{helpMaxResults}} must be between 1 to " + strconv.Itoa(EmbedLimitField))
	}
	if configData.BotOptions.YouTubeMaxResults > EmbedLimitField || configData.BotOptions.YouTubeMaxResults <= 0 {
		return errors.New("config:{botOptions:{youtubeMaxResults}} must be between 1 to " + strconv.Itoa(EmbedLimitField))
	}

	//Bot key checks
	if configData.BotOptions.UseDuckDuckGo && configData.BotKeys.DuckDuckGoAppName == "" {
		return errors.New("config:{botOptions:{useDuckDuckGo: true}} not permitted, config:{botKeys:{ddgAppName: \"\"}}")
	}
	if configData.BotOptions.UseImgur && configData.BotKeys.ImgurClientID == "" {
		return errors.New("config:{botOptions:{useImgur: true}} not permitted, config:{botKeys:{imgurClientID: \"\"}}")
	}
	if configData.BotOptions.UseSoundCloud && configData.BotKeys.SoundCloudAppVersion == "" {
		return errors.New("config:{botOptions:{useSoundCloud: true}} not permitted, config:{botKeys:{soundcloudAppVersion: \"\"}}")
	}
	if configData.BotOptions.UseSoundCloud && configData.BotKeys.SoundCloudClientID == "" {
		return errors.New("config:{botOptions:{useSoundCloud: true}} not permitted, config:{botKeys:{soundcloudClientID: \"\"}}")
	}
	if configData.BotOptions.UseWolframAlpha && configData.BotKeys.WolframAppID == "" {
		return errors.New("config:{botOptions:{useWolframAlpha: true}} not permitted, config:{botKeys:{wolframAppID: \"\"}}")
	}
	if configData.BotOptions.UseYouTube && configData.BotKeys.YouTubeAPIKey == "" {
		return errors.New("config:{botOptions:{useYouTube: true}} not permitted, config:{botKeys:{youtubeAPIKey: \"\"}}")
	}

	//Custom response checks
	for i, customResponse := range configData.CustomResponses {
		regexp, err := regexp.Compile(customResponse.Expression)
		if err != nil {
			return err
		}
		configData.CustomResponses[i].Regexp = regexp
	}
	return nil
}
