package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jdkato/prose/v2"
)

var posMap = map[string]string{
	"(":      "left round bracket",
	")":      "right round bracket",
	",":      "comma",
	":":      "colon",
	".":      "period",
	"''":     "closing quotation mark",
	"``":     "opening quotation mark",
	"#":      "number sign",
	"$":      "currency",
	"CC":     "coordinating conjunction",
	"CD":     "cardinal number",
	"DT":     "determiner",
	"FW":     "foreign word",
	"IN":     "subordinating or preposition conjunction",
	"JJ":     "adjective",
	"JJR":    "comparative adjective",
	"JJS":    "superlative adjective",
	"LS":     "list item marker",
	"MD":     "modal auxiliary verb",
	"NN":     "singular or mass noun",
	"NNP":    "proper singular noun",
	"NNPS":   "proper plural noun",
	"NNS":    "plural noun",
	"PDT":    "predeterminer",
	"POS":    "possessive ending",
	"PRP":    "personal pronoun",
	"PRP$":   "possessive pronoun",
	"RB":     "adverb",
	"RBR":    "comparative adverb",
	"RBS":    "superlative adverb",
	"RP":     "particle adverb",
	"SYM":    "symbol",
	"TO":     "infinitival to",
	"UH":     "interjection",
	"VB":     "base form verb",
	"VBD":    "past tense verb",
	"VBG":    "gerund or present participle verb",
	"VBN":    "past participle verb",
	"VBP":    "non-3rd person singular present",
	"VBZ":    "3rd person singular present",
	"WDT":    "wh-determiner",
	"WP":     "personal wh-pronoun",
	"WP$":    "possessive wh-pronoun",
	"WRB":    "wh-adverb",
	"PEOPLE": "person",
	"GPE":    "geographical/political entities",
}

func commandNLP(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	document, err := prose.NewDocument(strings.Join(args, " "))
	if err != nil {
		return NewErrorEmbed("Natural Language Processing - Error", "There was an error creating a document of your message.")
	}

	tokens := ""
	for _, token := range document.Tokens() {
		tokens += token.Text + " [" + posMap[token.Tag] + "]\n"
	}
	tokens = strings.TrimSuffix(tokens, "\n")

	sentences := ""
	for _, sentence := range document.Sentences() {
		sentences += sentence.Text + "\n"
	}
	sentences = strings.TrimSuffix(sentences, "\n")

	entities := ""
	for _, entity := range document.Entities() {
		entities += entity.Text + " [" + posMap[entity.Label] + "]\n"
	}
	entities = strings.TrimSuffix(entities, "\n")

	nlpEmbed := NewEmbed().
		SetTitle("Natural Language Processing").
		SetDescription("These are the results for the specified message.").
		SetFooter("Powered by Prose.").
		SetColor(0x1C1C1C)

	if tokens != "" {
		nlpEmbed.AddField("Part-of-Speech Tagging", tokens)
	}
	if sentences != "" {
		nlpEmbed.AddField("Sentences", sentences)
	}
	if entities != "" {
		nlpEmbed.AddField("Entities", entities)
	}

	return nlpEmbed.MessageEmbed
}
