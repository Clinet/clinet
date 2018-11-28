package main

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// CommandNLP holds a list of NLP commands to execute in order
type CommandNLP struct {
	Commands []*NLP //The slice of commands to execute in order
}

// NLP holds data related to a query command executable by any message system
//
// Command and Regex must always be specified.
// If RegexArguments is specified, it will be used for the process of pulling arguments.
// If RegexReplace is specified, it will be used for the process of ordering and modifying the arguments.
// Once a match succeeds and a string of arguments is pulled, it will then be split by spaces and passed
// along to the matching command specified.
//
/* Examples
* &CommandNLP{Command: "play", Regex: regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)")}
* &CommandNLP{Command: "queue", Regex: regexp.MustCompile("(?i)(?:.*?)(?:remove|delete)(?:\\s)queue(.*)"), Replace: "remove (${1})"}
 */
type NLP struct {
	Command        string         //The name of a command to execute when this NLP command is triggered
	ArgPrefix      string         //The prefix to apply to pulled arguments
	ArgSuffix      string         //The suffix to apply to pulled arguments
	Regex          *regexp.Regexp //The regular expression to match against
	RegexArguments *regexp.Regexp //The regular expression to pull arguments with
	RegexReplace   string         //The replacement scheme to follow for pulling arguments
}

func initNLPCommands() {
	addNLPCommand(nlpNew("spotify", "search", "", regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)(?:\\s)(?:from Spotify|on Spotify)(?:.*)"), nil, "${1}"),
		nlpNew("spotify", "play", "", regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)(?:\\s)(?:from Spotify|on Spotify)(?:.*)"), nil, "1"))
	addNLPCommand(nlpNew("play", "", "", regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)"), nil, ""))
	addNLPCommand(nlpNew("queue", "remove", "", regexp.MustCompile("(?i)(?:.*?)(?:.*?)(?:remove|delete)(?:.*)(\\b\\d+)(?:.*)(?:queue)(?:.*)"), regexp.MustCompile("(\\b\\d+)"), ""))
	addNLPCommand(nlpNew("remind", "", "", regexp.MustCompile("(?i)(.*?)(?:.*?)(remind me|set a reminder)(.*)"), nil, ""))
}

func nlpNew(command, argPrefix, argSuffix string, regex, regexArguments *regexp.Regexp, regexReplace string) *NLP {
	return &NLP{Command: command, ArgPrefix: argPrefix, ArgSuffix: argSuffix, Regex: regex, RegexArguments: regexArguments, RegexReplace: regexReplace}
}

func addNLPCommand(nlp ...*NLP) {
	botData.NLPCommands = append(botData.NLPCommands, &CommandNLP{Commands: nlp})
}

func callNLP(message string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for _, command := range botData.NLPCommands {
		for i := 0; i < len(command.Commands); i++ {
			nlp := command.Commands[i]

			if !nlp.Regex.MatchString(message) {
				continue
			}

			if _, exists := botData.Commands[nlp.Command]; !exists {
				continue
			}

			env.Command = nlp.Command

			matches := make([]string, 0)
			if nlp.RegexArguments != nil {
				if nlp.RegexReplace != "" {
					replaced := nlp.RegexArguments.ReplaceAllString(message, nlp.RegexReplace)
					matches = strings.Split(replaced, " ")
				} else {
					matches = nlp.RegexArguments.FindAllString(message, -1)
				}
			} else {
				if nlp.RegexReplace != "" {
					replaced := nlp.Regex.ReplaceAllString(message, nlp.RegexReplace)
					matches = strings.Split(replaced, " ")
				} else {
					matches = nlp.Regex.FindAllString(message, -1)
				}
			}

			if nlp.ArgPrefix != "" {
				prefixes := strings.Split(nlp.ArgPrefix, " ")
				matches = append(prefixes, matches...)
			}
			if nlp.ArgSuffix != "" {
				suffixes := strings.Split(nlp.ArgSuffix, " ")
				matches = append(matches, suffixes...)
			}
			Debug.Println("Matches:", matches)

			embed := callCommand(nlp.Command, matches, env)
			if i == (len(command.Commands) - 1) {
				return embed
			} else {
				if strings.Contains(embed.Title, "Error") {
					return embed
				}
			}
		}
	}
	return nil
}
