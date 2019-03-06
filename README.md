# clinet

### A Discord bot intended for assistance and control within your guilds

[![Go Report Card](https://goreportcard.com/badge/github.com/JoshuaDoes/clinet)](https://goreportcard.com/report/github.com/JoshuaDoes/clinet)

[![Discord Bots](https://discordbots.org/api/widget/374546169755598849.svg)](https://discordbots.org/bot/374546169755598849)

----

## Using the official live version of Clinet (to stay up to date on fixes and features)

1. Click on [this link](https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot) to invite `Clinet` into your Discord server.
    1. The `Administrator` permission is not required. The invite link above
    only requests the permission to handle all of Clinet's advanced features,
    saving time from going through and adding missing permissions. Should you
    still wish to manage these permissions yourself, do not grant the bot the
    `Administrator` permission. At a later date, the bot will give errors for
    permissions it does not have when asked to do something, and should sending
    the error message fail it will warn the server owner in a DM.
2. (Optional) Join the official [Clinet Discord server](https://discord.gg/zx2ns2J) to get updates on the latest features and changes Clinet has to offer, get help with issues you may be having, and even use bot testing channels to test-drive the bot before deciding if it's right for your community!
3. Enjoy!

## What does it do?

After the `Clinet` bot is invited to your Discord server, it will immediately
begin listening for certain keywords within conversations to trigger certain
events.

`Clinet` will listen for a message that prefixes a query with a mention of it to detect when it is
being queried to do something or answer a question. It begins by checking a list of RegEx expressions
stored in the bot configuration to look for configurable responses that bot hosters can set
for their specific instance of the bot. If nothing is found, then it continues on to check for
hard-coded natural language queries to trigger various commands. If that fails, it again continues on to query
DuckDuckGo's Instant Answers API for a possible response. If DuckDuckGo comes up short, it
finally queries Wolfram|Alpha as a last resort (as Wolfram|Alpha's API services are limited for
non-paying developers). Should no responses be found from any of the three sources, Clinet tells
the user that there was an error finding a response for the given query.

Additionally, `Clinet` supports various functionalities not available in question-response queries
using commands. Commands are prefixed by default using `cli$` and sometimes take parameters to
control the output and action of the command. Server owners can change the prefix for their servers
after inviting Clinet to the server.

Finally, `Clinet` has some very useful message management features proven to be successful in the
servers it resides in. If you send a query and make a mistake, for example if you misspell a word,
forget to tag the bot, or forget some form of required punctuation, you don't have to send a whole
new message - just edit the previous message with the fix and Clinet happily responds to the updated
message. Adding onto this, you can even edit a message that had a successful response - Clinet will
happily edit its response message with the new response to the updated query. Lastly, if you delete
your original query message, Clinet will help with the chat cleanup and delete its response message.
These message management features work for both question queries and commands, and soon will be
interwoven into music playback commands.

## Commands

The default configuration of Clinet uses `cli$` as the command prefix. This can be configured in
the bot configuration for bot hosts, however remains a unique command prefix that should never
interfere with another bot's default command prefix.

All of Clinet's commands respond using a rich embed with all fields inlined to save chat screen
estate on desktop and web versions of Discord while maintaining a clean output everywhere.

For a list of available commands, use the `cli$help` command in a server with Clinet.

----

## Rolling your own locally

In order to run `Clinet` locally, you must have already installed a working Golang
environment on your development system. Clinet is currently built using Golang `1.11`,
and earlier versions of Go are not guaranteed to be supported at this time.

### Fetching Clinet and dependencies

Run `go get github.com/JoshuaDoes/clinet` and watch the magic happen!

### Building

`Clinet` is built using a compiler wrapper known as `govvv`, and opts to use an
altered version to support additional things. govvv acts as a git status injector
for the output compiled binary, taking current statuses of the git repo Clinet is
in and injecting them into uninitialized strings in the main source file to be used
in the command `cli$version`. Simply follow the instructions on the [govvv](https://github.com/JoshuaDoes/govvv) repo page
to learn how to install and use it, then run `govvv build` in the Clinet repo
directory.

### Acquiring necessary API keys

Clinet's functionality relies on a set of different API keys and access tokens, and without them sports less features to interact with and use. The official bot has all of these already, but if you're looking to roll your own instance of the bot you'll need to acquire these on your own (an exercise left up to you).

| Services | Requirements |
| -------- | ------------ |
| Wolfram\|Alpha | App ID |
| DuckDuckGo | App name (can be anything) |
| YouTube | Search API key |
| Imgur | Client ID |
| SoundCloud | Client ID and app version |

### Writing the configuration

`Clinet` stores its configuration in a file named `config.json` using the JSON data
structure. It has a number of configurable variables and will always globally
override a server's settings if it disables a feature. Passing the command line
argument `-config test.json` will instead load the bot configuration from `test.json`.

An example of an empty configuration file can be found in `config.example.json`.

Most of the configuration options should be self-explanatory, but here's some explanations for a few of the less guessable ones:

| Variable | Description |
| -------- | ----------- |
| `botToken` | The token of the bot account Clinet should log into. Can be acquired by [creating an application and then declaring it as a bot user](https://discordapp.com/developers/applications/me/create) and/or [selecting a pre-existing bot user application and acquiring the bot token under the `APP BOT USER` section](https://discordapp.com/developers/applications/me). |
| `botOwnerID` | The user ID of the bot owner. Can be acquired by enabling developer mode on Discord, right clicking your user in a server's user list, and clicking `Copy ID`. If Clinet crashes and recovers from the crash, the error and a full stack trace will be directly messaged to whatever user this option is set to. |
| `sendOwnerStackTraces` | If this is set to true, the bot owner specified in `botOwnerID` will receive crash reports when Clinet recovers from a crash. |
| `botOptions` -> `maxPingCount` | The amount of ping messages to send to Discord to test the ping average when using the `ping` command. This has a maximum of 5 to prevent inconsistent results due to Discord's API ratelimits, whereas the example configuration sets this to 4 so the results embed isn't stuck because of the API rate limit and can send immediately.
| `botOptions` -> `sendTypingEvent` | Whether or not to send a typing notification in a channel containing a query or command for Clinet to respond to. Helpful for queries or commands that take a little longer than usual to respond to so users know the bot isn't broken. |
| `botOptions` -> `wolframDeniedPods` | An array of pod titles to skip over when creating a list of responses to use in a rich embed response from a Wolfram\|Alpha query. The default list is highly recommended for bot hosters concerned with the privacy of the bot's host location. |
| `botOptions` -> `youtubeMaxResults` | The total amount of results to display per page for YouTube searches via the `cli$youtube search` command. Maximum of 253. |
| `debugMode` | Debug mode enables various console debugging features, such as chat output and other detailed information about what Clinet is up to. |
| `customResponses` | Stored as objects in an array, custom responses are exactly what the name depicts. Each object contains an `expression` variable, which stores a valid regular expression, and a `responses` array, which itself contains objects randomly selected by the main program for different `responseEmbed` responses each time the custom response is queried. Alternatively, you can specify a `cmdResponses` array, which also contains objects randomly selected by the main program for different `commandName` commands to execute with the arguments in `args`. Command responses are direct executions of available commands in Clinet with any given parameters. |
| `customStatuses` | Stored as objects in an array, custom statuses are used to set the bot's presence. Each object contains a `type` variable, which stores integers 0, 1, and 2, which are "Playing", "Listening to", and "Streaming" respectively, and a `status` variable, which stores the status text to use. If the type is set to 2, you can also set a `url` variable to use as the stream URL. |

The configuration file by default will never be included in git commits, as declared by `.gitignore`. This is to prevent accidental leakage of API keys and bot tokens.

### Running `Clinet`

Finally, to run Clinet, simply type `./clinet` in your terminal/shell or `.\clinet.exe` in your command prompt. If everything goes well, you can find your bot user application and generate an OAuth2 URL to invite the bot into various servers in which you have the `Administrator` permission of.

### Debug mode

To start Clinet with debug mode enabled, simply type `./clinet -debug true` in your terminal/shell or `.\clinet.exe -debug true` in your command prompt. To toggle debug mode on-the-fly, type `cli$debug` in any channel Clinet can read from.

When running Clinet in debug mode, a surplus of debug logging will be outputted to your terminal's STDOUT pipe. This includes debugging information reported by discordgo and the various happenings within Clinet, including the commands ran by other users and the resulting responses generated by Clinet (including embeds).

### Panic recovery

If Clinet ever crashes from a panic, custom-made panic recovery will save the crash message to `crash.txt` and the stack trace to `stacktrace.txt` in the bot's working directory. When Clinet is next started up, it will send the crash message and the file of the stack trace to the user specified in the configuration option `botOwnerID` and proceed to delete the two files.

Running Clinet by itself will spawn a "master" process with a few small jobs: Spawning a "bot" process, restarting the "bot" process if it exits for any reason, and closing the "bot" process if the "master" process ever exits for any reason. This is to ensure that, even if the "bot" process crashes, Clinet can continue running and instantly report the crash to the user specified in the configuration option `botOwnerID`.

### States

If you close Clinet after running it long enough for it to merely exist on Discord, you'll notice a new folder called `state`. This folder contains "states" of various structs within Clinet's memory, stored in pretty-printed JSON format. Upon reopening Clinet, these state files are then loaded into memory so Clinet can (for the most part) return to its original "state" before it was closed. States were added as helpers to panic recovery so users can continue with what they were doing, and will be replaced with a proper database engine at a later date.

### Updating

If you want to keep Clinet up to date without manually running ``go get github.com/JoshuaDoes/clinet``, ``go build github.com/JoshuaDoes/clinet``, and running Clinet again, you have the full ability to do so! Make sure your Discord user ID is specified as the bot owner in Clinet's configuration and run `cli$update` whenever a new commit is pushed. And if you need to make sure it works without waiting on a new update, run `cli$update force`.

----

## Support
For help and support with Clinet, visit the [Clinet Discord server](https://discord.gg/qkbKEWT) and ask for an online developer.

## License
The source code for Clinet is released under the MIT License. See LICENSE for more details.

## Donations
All donations are highly appreciated. They help me pay for the server costs to keep Clinet running and even help point my attention span to Clinet to fix issues and implement newer and better features!

[![Donate](https://img.shields.io/badge/Donate-PayPal-green.svg)](https://paypal.me/JoshuaDoes)
