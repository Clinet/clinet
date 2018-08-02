package main

import (
	"sort"
	
	"github.com/bwmarrin/discordgo"
)

func commandAbout(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - About").
		SetDescription(botData.BotName+" is a Discord bot written in Google's Go programming language, intended for conversation and fact-based queries.").
		AddField("How can I use "+botData.BotName+" in my server?", "Simply open the Invite Link at the end of this message and follow the on-screen instructions.").
		AddField("How can I help keep "+botData.BotName+" running?", "The best ways to help keep "+botData.BotName+" running are to either donate using the Donation Link or contribute to the source code using the Source Code Link, both at the end of this message.").
		AddField("How can I use "+botData.BotName+"?", "There are many ways to make use of "+botData.BotName+".\n1) Type ``cli$help`` and try using some of the available commands.\n2) Ask "+botData.BotName+" a question, ex: ``@"+botData.BotName+"#1823, what time is it?`` or ``@"+botData.BotName+"#1823, what is DiscordApp?``.").
		AddField("Where can I join the "+botData.BotName+" Discord server?", "If you would like to get help and support with "+botData.BotName+" or experiment with the latest and greatest of "+botData.BotName+", use the Discord Server Invite Link at the end of this message.").
		AddField("Invite Link", "https://discordapp.com/api/oauth2/authorize?client_id=374546169755598849&permissions=8&scope=bot").
		AddField("Donation Link", "https://www.paypal.me/JoshuaDoes").
		AddField("Source Code Link", "https://github.com/JoshuaDoes/clinet-discord/").
		AddField("Discord Server Invite Link", "https://discord.gg/qkbKEWT").
		SetColor(0x1C1C1C).MessageEmbed
}
func commandHelp(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	//First see if help text is being requested for a particular command
	if len(args) > 0 {
		if command, exists := botData.Commands[args[0]]; exists {
			if command.IsAlternateOf != "" {
				if commandAlternate, exists := botData.Commands[command.IsAlternateOf]; exists {
					command = commandAlternate
				} else {
					return nil
				}
			}
			return getCommandUsage(args[0], "Help for **"+args[0]+"**")
		}
	}

	//Before we fetch help text data, we need to have an alphabetical listing of commands
	var commandMapKeys []string
	for commandMapKey := range botData.Commands {
		commandMapKeys = append(commandMapKeys, commandMapKey)
	}
	sort.Strings(commandMapKeys)

	//Create a dynamic list of fields for the help embed
	commandFields := []*discordgo.MessageEmbedField{}

	//Iterate over the alphabetically sorted command list and add each listed command to the help embed field list
	for _, commandName := range commandMapKeys {
		command := botData.Commands[commandName]
		if command.IsAlternateOf == "" {
			if permissionsAllowed, _ := MemberHasPermission(botData.DiscordSession, env.Guild.ID, env.User.ID, discordgo.PermissionAdministrator|command.RequiredPermissions); permissionsAllowed || command.RequiredPermissions == 0 {
				commandField := &discordgo.MessageEmbedField{Name: botData.CommandPrefix + commandName, Value: command.HelpText, Inline: true}
				commandFields = append(commandFields, commandField)
			}
		}
	}

	//Create the help embed and give it the command list
	helpEmbed := NewEmbed().
		SetTitle(botData.BotName + " - Help").
		SetDescription("A list of commands you have permission to use.").
		SetColor(0xFAFAFA).MessageEmbed
	helpEmbed.Fields = commandFields

	//Return the help embed to the caller
	return helpEmbed
}
func commandVersion(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Version").
		AddField("Build ID", BuildID).
		AddField("Build Date", BuildDate).
		AddField("Latest Development", GitCommitMsg).
		AddField("GitHub Commit URL", GitHubCommitURL).
		AddField("Golang Version", GolangVersion).
		SetColor(0x1C1C1C).MessageEmbed
}
func commandCredits(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	return NewEmbed().
		SetTitle(botData.BotName+" - Credits").
		AddField("Bot Development", "- JoshuaDoes (2018)").
		AddField("Programming Language", "- Golang").
		AddField("Golang Libraries", "- https://github.com/bwmarrin/discordgo\n"+
			"- https://github.com/disintegration/gift\n"+
			"- https://github.com/JoshuaDoes/duckduckgolang\n"+
			"- https://github.com/google/go-github/github\n"+
			"- https://github.com/jonas747/dca\n"+
			"- https://github.com/JoshuaDoes/go-soundcloud\n"+
			"- https://github.com/JoshuaDoes/go-wolfram\n"+
			"- https://github.com/koffeinsource/go-imgur\n"+
			"- https://github.com/koffeinsource/go-klogger\n"+
			"- https://github.com/nishanths/go-xkcd\n"+
			"- https://github.com/paked/configure\n"+
			"- https://github.com/robfig/cron\n"+
			"- https://github.com/rylio/ytdl\n"+
			"- https://google.golang.org/api/googleapi/transport\n"+
			"- https://google.golang.org/api/youtube/v3").
		AddField("Icon Design", "- thejsa").
		AddField("Source Code", "- https://github.com/JoshuaDoes/clinet-discord").
		SetColor(0x1C1C1C).MessageEmbed
}