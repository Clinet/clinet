# clinet-discord

### A Discord bot intended for conversation and fact-based queries

----

## Using the official up-to-date version of Clinet

1. Click on [this link](https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot) to invite `Clinet` into your Discord server
    1. The 'Administrator' permission is not required. The invite link above
    will create a 'Clinet' role for the bot, and in the future some advanced
	features may require the 'Administrator' permission.
2. Enjoy!

## How does it work?

After the `Clinet` bot is invited to your Discord server, it will immediately
begin listening for certain keywords within conversations to trigger certain
events.

Currently, `Clinet` will listen for its name and a question mark (?) at the end of
a message to detect when it is being queried a question. It will then query
Wolfram|Alpha with the question and then send a message with the response to the
text channel it was queried in.

**Note: Clinet is not yet ready for use in public servers. You have been warned.**

## Commands

```
cli$play (url) - Plays the specified URL in a voice channel via YouTube-DL
cli$youtube search (query) - Searches for the queried video and plays it in a voice channel via YouTube-DL
cli$stop - Stops the currently playing audio
cli$leave - Leaves the current voice chat
```

~~Here's a [list of supported sites](https://rg3.github.io/youtube-dl/supportedsites.html) under YouTube-DL.~~
**Note: Under the current setup, Clinet is only prepared to play audio from YouTube.**

----

## Rolling your own locally
 
In order to run `Clinet` locally, you will need to edit the JSON configuration file
`config.json` with the appropriate values.

In the below configuration template, use the following keymap:
```
$BotToken$ - The bot token assigned to your bot application by Discord
$BotName$ - The name of your bot; used to detect queries
$BotPrefix$ - The prefix to use for various commands, ex. "cli$" for "cli$play"
$WolframAppID$ - The AppID for your Wolfram|Alpha account
```

**Configuration template:**
```JSON
{
	"botToken": "$BotToken$",
	"botName": "$BotName$",
	"botPrefix": "$BotPrefix$",
	"wolframAppID": "$WolframAppID"
}
```

----

## Dependencies

| [configure](https://github.com/paked/configure) |
| [DiscordGo](https://github.com/bwmarrin/discordgo) |
| [go-wolfram](https://github.com/JoshuaDoes/go-wolfram) |
| [dca](https://github.com/jonas747/dca) |
| [ytdl](https://github.com/rylio/ytdl) |

## License
The source code for Clinet is released under the MIT License. See LICENSE for more details.