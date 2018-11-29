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
// If RegexBadMatch is specified, it will be used for the process of checking for bad matches. If the
// subject message matches this expression, the NLP handler will bail.
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
	RegexBadMatch  *regexp.Regexp //The regular expression to test bad matches against
	RegexArguments *regexp.Regexp //The regular expression to pull arguments with
	RegexReplace   string         //The replacement scheme to follow for pulling arguments
}

func initNLPCommands() {
	//@Clinet Play Dance Gavin Dance on Spotify
	addNLPCommand(nlpNew("spotify", "search", "", regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)(?:\\s)(?:from Spotify|on Spotify)(?:.*)"), nil, nil, "${1}"),
		nlpNew("spotify", "play", "", regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*)(?:\\s)(?:from Spotify|on Spotify)(?:.*)"), nil, nil, "1"))

	//@Clinet Play Dance Gavin Dance
	addNLPCommand(nlpNew("play", "", "", regexp.MustCompile("(?i)(?:.*?)(?:play|listen to)(?:\\s)(.*?)"), nil, nil, ""))

	//@Clinet Remove the 1st entry from the queue
	addNLPCommand(nlpNew("queue", "remove", "", regexp.MustCompile("(?i)(?:.*?)(?:remove|delete)(?:.*?)(\\b\\d+)(?:.*)(?:queue)(?:.*)"), nil, regexp.MustCompile("(\\b\\d+)"), ""))

	//@Clinet Can the queue be cleared?
	addNLPCommand(nlpNew("queue", "clear", "", regexp.MustCompile("(?i)(?:.*?)(?:queue)(?:.*?)(?:clear)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet Clear the queue
	addNLPCommand(nlpNew("queue", "clear", "", regexp.MustCompile("(?i)(?:.*?)(?:clear)(?:.*?)(?:queue)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet List the queue entries
	addNLPCommand(nlpNew("queue", "", "", regexp.MustCompile("(?i)(?:.*?)(?:queue)(?:.*?)"), regexp.MustCompile("(?i)(?:.*?)(?:remove|delete)(?:.*?)"), nil, "${1}"))

	//@Clinet List the 1st page of the queue
	addNLPCommand(nlpNew("queue", "", "", regexp.MustCompile("(?i)(?:.*?)(\\d+)(?:.*?)(?:queue)(?:.*?)"), regexp.MustCompile("(?i)(?:.*?)(?:remove|delete)(?:.*?)"), regexp.MustCompile("(\\d+)"), ""))

	//@Clinet Skip this song
	addNLPCommand(nlpNew("skip", "", "", regexp.MustCompile("(?i)(?:.*?)(?:skip|next)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet Stop the playback
	addNLPCommand(nlpNew("stop", "", "", regexp.MustCompile("(?i)(?:.*?)(?:stop)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet Pause the song
	addNLPCommand(nlpNew("pause", "", "", regexp.MustCompile("(?i)(?:.*?)(?:pause)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet Resume the playback
	addNLPCommand(nlpNew("resume", "", "", regexp.MustCompile("(?i)(?:.*?)(?:resume)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet What are the lyrics for this song?
	addNLPCommand(nlpNew("lyrics", "", "", regexp.MustCompile("(?i)(?:.*?)(?:lyrics)(?:.*?)"), nil, nil, "${1}"))

	//@Clinet Set a reminder to add some cool stuff in 1 hour
	addNLPCommand(nlpNew("remind", "", "", regexp.MustCompile("(?i)(.*?)(?:.*?)(remind me|set a reminder)(.*?)"), nil, nil, ""))

	//@Clinet What are my reminders?
	addNLPCommand(nlpNew("remind", "list", "", regexp.MustCompile("(?i)(?:.*?)(?:reminders)(?:.*?)"), nil, nil, "${1}"))
}

func nlpNew(command, argPrefix, argSuffix string, regex, regexBadMatch, regexArguments *regexp.Regexp, regexReplace string) *NLP {
	return &NLP{Command: command, ArgPrefix: argPrefix, ArgSuffix: argSuffix, Regex: regex, RegexBadMatch: regexBadMatch, RegexArguments: regexArguments, RegexReplace: regexReplace}
}

func addNLPCommand(nlp ...*NLP) {
	botData.NLPCommands = append(botData.NLPCommands, &CommandNLP{Commands: nlp})
}

func callNLP(message string, env *CommandEnvironment) *discordgo.MessageEmbed {
	for i, command := range botData.NLPCommands {
		for j := 0; j < len(command.Commands); j++ {
			Debug.Printf("Testing NLP %d, command %d...", i, j)

			nlp := command.Commands[j]

			if nlp.Regex == nil {
				break
			}

			if !nlp.Regex.MatchString(message) {
				break
			}

			if nlp.RegexBadMatch != nil {
				if nlp.RegexBadMatch.MatchString(message) {
					break
				}
			}

			if _, exists := botData.Commands[nlp.Command]; !exists {
				break
			}

			env.Command = nlp.Command

			matches := make([]string, 0)
			if nlp.RegexArguments != nil {
				Debug.Println("RegexArguments: True")
				if nlp.RegexReplace != "" {
					Debug.Println("RegexReplace: True")
					replaced := nlp.RegexArguments.ReplaceAllString(message, nlp.RegexReplace)
					matches = strings.Split(replaced, " ")
				} else {
					Debug.Println("RegexReplace: False")
					matches = nlp.RegexArguments.FindAllString(message, -1)
				}
			} else {
				Debug.Println("RegexArguments: False")
				if nlp.RegexReplace != "" {
					Debug.Println("RegexReplace: True")
					replaced := nlp.Regex.ReplaceAllString(message, nlp.RegexReplace)
					matches = strings.Split(replaced, " ")
				} else {
					Debug.Println("RegexReplace: False")
					Debug.Println(nlp.RegexArguments)
					Debug.Println(nlp.RegexReplace)
					matches = nlp.Regex.FindAllString(message, -1)
				}
			}

			if nlp.ArgPrefix != "" {
				Debug.Println("Prefix:", nlp.ArgPrefix)
				prefixes := strings.Split(nlp.ArgPrefix, " ")
				matches = append(prefixes, matches...)
			}
			if nlp.ArgSuffix != "" {
				Debug.Println("Suffix:", nlp.ArgSuffix)
				suffixes := strings.Split(nlp.ArgSuffix, " ")
				matches = append(matches, suffixes...)
			}
			Debug.Printf("Matches (%d): %v", len(matches), matches)

			if len(matches) == 1 && matches[0] == "" {
				matches = nil
			}

			embed := callCommand(nlp.Command, matches, env)
			if j == (len(command.Commands) - 1) {
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
