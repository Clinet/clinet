# clinet-discord

### A Discord bot intended for conversation and fact-based queries

----

## Using the official up-to-date version of Clinet

1. Click on [this link](https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot) to invite `Clinet` into your Discord server.
    1. The 'Administrator' permission is not required. The invite link above
    will create a 'Clinet' role for the bot, and in the future some advanced
	features may require the 'Administrator' permission.
2. (Optional) Join the official [Clinet Discord server](https://discord.gg/qkbKEWT) to get updates on the latest features and changes and also test-drive Clinet before using it on your own server.
3. Enjoy!

## How does it work?

After the `Clinet` bot is invited to your Discord server, it will immediately
begin listening for certain keywords within conversations to trigger certain
events.

Currently, `Clinet` will listen for its name and a question mark (?) at the end of
a message to detect when it is being queried with a question. It will then query
DuckDuckGo with the question and then send a message with the response to the
text channel it was queried in. Though, if DuckDuckGo fails, it will next attempt
to use Wolfram|Alpha. If both fail, Clinet tells the user that there was an error.

**Note: Clinet is not yet ready for use in public servers. You have been warned.**

## Commands

```
cli$help - Lists available commands.
cli$about - Displays information about Clinet and how to use it.
cli$roll - Rolls a dice.
cli$doubleroll - Rolls two die.
cli$coinflip - Flips a coin.
cli$xkcd (comic number|random|latest) - Displays an xkcd comic depending on the request type or comic number.
cli$imgur (url) - Displays info about the specified Imgur image, album, gallery image, or gallery album.
cli$play (url/YouTube search query) - lays either the first result from the specified YouTube search query or the specified YouTube/direct audio URL in the user's current voice channel.
cli$stop - Stops the currently playing audio.
cli$skip - Stops the currently playing audio, and, if available, attempts to play the next audio in the queue.
cli$queue - Lists all entries in the queue.
cli$clear - Clears the current queue.
cli$leave - Leaves the current voice channel.
```

~~Here's a [list of supported sites](https://rg3.github.io/youtube-dl/supportedsites.html) under YouTube-DL.~~
**Note: Under the current setup, Clinet is only prepared to play audio from YouTube.**

----

## Rolling your own locally
 
In order to run `Clinet` locally, you will need to create a JSON configuration
file called `config.json` with the appropriate values.

In the below configuration template, use the following keymap:
```
$BotToken$ - The bot token assigned to your bot application by Discord (string)
$BotName$ - The name of your bot; used to detect queries (string)
$BotPrefix$ - The prefix to use for various commands, ex. "cli$" for "cli$play" (string)
$WolframAppID$ - The App ID for your Wolfram|Alpha account (string)
$DuckDuckGoAppName$ - The app name to use for DuckDuckGo Instant Answer API queries (string)
$YouTubeAPIKey$ - The API key to use for YouTube API v3 (string)
$ImgurClientID$ - The Client ID to use for Imgur API info requests (string)
$DebugMode$ - Whether or not to enable debug mode (bool, optional)
```

**Configuration template:**
```JSON
{
	"botToken": "$BotToken$",
	"botName": "$BotName$",
	"botPrefix": "$BotPrefix$",
	"wolframAppID": "$WolframAppID",
	"ddgAppName": "$DuckDuckGoAppName$",
	"youtubeAPIKey": "$YouTubeAPIKey$",
	"imgurClientID": "$ImgurClientID$",
	"debugMode": $DebugMode$
}
```

In addition, you may use RegEx to add custom responses for `Clinet` to use.
Custom responses are checked before `Clinet` attempts to query online sources
for answers, so you may override answers or even add your own using custom
responses.

To add custom responses to `Clinet`, you must edit your `config.json` file
to include them, like so:

```JSON
{
	...
	"debugMode": true,
	"customResponses": [
		{
			"regex": "(.*)what's your name(.*)",
			"response": "My name is Clinet."
		},
		{
			"regex": "(.*)who created you(.*)",
			"response": "I was created by JoshuaDoes."
		}
	]
}
```

**Note: Before custom responses are checked, the query still runs through the
usual sanitization methods. When creating your custom responses, craft the regex
with the prefix `Clinet` and the suffix `?` in mind.**

----

## Dependencies

| [configure](https://github.com/paked/configure) |
| [DiscordGo](https://github.com/bwmarrin/discordgo) |
| [go-wolfram](https://github.com/JoshuaDoes/go-wolfram) |
| [dca](https://github.com/jonas747/dca) |
| [ytdl](https://github.com/rylio/ytdl) |
| [duckduckgolang](https://github.com/JoshuaDoes/duckduckgolang) |
| [go-imgur](https://github.com/koffeinsource/go-imgur) |

## License
The source code for Clinet is released under the MIT License. See LICENSE for more details.

## Donations
All donations are appreciated and help pay for the costs of the server Clinet is officially hosted on. Even if it's not much, it helps a lot in the long run!
You can find the donation link here: [Donation Link](https://paypal.me/JoshuaDoes)