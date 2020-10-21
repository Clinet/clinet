package main

import (
	"errors"
	"strings"
	"time"

	"github.com/bregydoc/gtranslate"
)

//translate translates a message from the detected language to the specified language
func translate(toLang, msg string) (translated string, err error) {
	return translateFrom("auto", toLang, msg)
}

//translateFrom translates a message from the specified language to the specified language
func translateFrom(fromLang, toLang, msg string) (translated string, err error) {
	if fromLang == "" || toLang == "" {
		return msg + " (translation failed)", errors.New("translator: language not specified")
	}
	if toLang == "auto" {
		return msg + " (translation failed)", errors.New("translator: cannot detect target language")
	}
	fromLang = getLanguageCode(fromLang)
	if fromLang == "" {
		return msg + " (translation failed)", errors.New("translator: unknown language " + fromLang)
	}
	toLang = getLanguageCode(toLang)
	if toLang == "" {
		return msg + " (translation failed)", errors.New("translator: unknown language " + toLang)
	}
	if fromLang == toLang {
		return msg, nil
	}

	translated, err = gtranslate.TranslateWithParams(msg, gtranslate.TranslationParams{From: fromLang, To: toLang, Tries: 5, Delay: time.Second * 1, GoogleHost: "google.com"})
	if err != nil {
		return msg + " (translation failed)", err
	}
	return translated, nil
}

func getLanguageCode(language string) (langCode string) {
	language = strings.ToLower(language)

	switch language {
	case "auto":
		return "auto"
	case "af", "afrikaans":
		return "af"
	case "sq", "albanian":
		return "sq"
	case "am", "amharic":
		return "am"
	case "ar", "arabic":
		return "ar"
	case "hy", "armenian":
		return "hy"
	case "az", "azerbaijani":
		return "az"
	case "eu", "basque":
		return "eu"
	case "be", "belarusian":
		return "be"
	case "bn", "bengali":
		return "bn"
	case "bs", "bosnian":
		return "bs"
	case "bg", "bulgarian":
		return "bg"
	case "ca", "catalan":
		return "ca"
	case "ceb", "cebuano":
		return "ceb"
	case "zh", "zh-CN", "chinese", "chinesesimpl", "chinesesimplified", "chinese(simplified)":
		return "zh"
	case "zh-tw", "chinesetrad", "chinesetradition", "chinesetraditional", "chinese(traditional)":
		return "zh-TW"
	case "co", "corsican":
		return "co"
	case "hr", "croatian":
		return "hr"
	case "cs", "czech":
		return "cs"
	case "da", "danish":
		return "da"
	case "nl", "dutch":
		return "nl"
	case "en", "en-US", "en-CA", "en-UK", "en-AU", "english":
		return "en"
	case "eo", "esperanto":
		return "eo"
	case "et", "estonian":
		return "et"
	case "fi", "finnish":
		return "fi"
	case "fr", "french":
		return "fr"
	case "fy", "frisian":
		return "fy"
	case "gl", "galician":
		return "gl"
	case "ka", "georgian":
		return "ka"
	case "de", "german":
		return "de"
	case "el", "greek":
		return "el"
	case "gu", "gujarati":
		return "gu"
	case "ht", "haitian", "haitiancreole", "creole":
		return "ht"
	case "ha", "hausa":
		return "ha"
	case "haw", "hawaiian":
		return "haw"
	case "he", "iw", "hebrew":
		return "he"
	case "hi", "hindi":
		return "hi"
	case "hmn", "hmong":
		return "hmn"
	case "hu", "hungarian":
		return "hu"
	case "is", "icelandic":
		return "is"
	case "ig", "igbo":
		return "ig"
	case "id", "indonesian":
		return "id"
	case "ga", "irish":
		return "ga"
	case "it", "italian":
		return "it"
	case "ja", "japanese", "weeaboo", "weeb":
		return "ja"
	case "jv", "javanese":
		return "jv"
	case "kn", "kannada":
		return "kn"
	case "kk", "kazakh":
		return "kk"
	case "km", "khmer":
		return "km"
	case "rw", "kinyarwanda":
		return "rw"
	case "ko", "korean":
		return "ko"
	case "ku", "kurdish":
		return "ku"
	case "ky", "kyrgyz":
		return "ky"
	case "lo", "lao":
		return "lo"
	case "la", "latin":
		return "la"
	case "lv", "latvian":
		return "lv"
	case "lt", "lithuanian":
		return "lt"
	case "lb", "luxembourgish":
		return "lb"
	case "mk", "macedonian":
		return "mk"
	case "mg", "malagasy":
		return "mg"
	case "ms", "malay":
		return "ms"
	case "ml", "malayalam":
		return "ml"
	case "mt", "maltese":
		return "mt"
	case "mi", "maori":
		return "mi"
	case "mr", "marathi":
		return "mr"
	case "mn", "mongolian":
		return "mn"
	case "my", "myanmar", "burmese", "myanmarburmese", "myanmar(burmese)":
		return "my"
	case "ne", "nepali":
		return "ne"
	case "no", "norwegian":
		return "no"
	case "ny", "nyanja", "chichewa", "nyanjachichewa", "nyanja(chichewa)":
		return "ny"
	case "or", "odia", "oriya", "odiaoriya", "odia(oriya)":
		return "or"
	case "ps", "pashto":
		return "ps"
	case "fa", "persian":
		return "fa"
	case "pl", "polish":
		return "pl"
	case "pt", "portuguese", "portugal", "brazil":
		return "pt"
	case "pa", "punjabi":
		return "pa"
	case "ro", "romanian":
		return "ro"
	case "ru", "russian":
		return "ru"
	case "sm", "samoan":
		return "sm"
	case "gd", "scots", "gaelic", "scotsgaelic":
		return "gd"
	case "sr", "serbian":
		return "sr"
	case "st", "sesotho":
		return "st"
	case "sn", "shona":
		return "sn"
	case "sd", "sindhi":
		return "sd"
	case "si", "sinhala", "sinhalese", "sinhala(sinhalese)":
		return "si"
	case "sk", "slovak":
		return "sk"
	case "sl", "slovenian":
		return "sl"
	case "so", "somali":
		return "so"
	case "es", "spanish":
		return "es"
	case "su", "sundanese":
		return "su"
	case "sw", "swahili":
		return "sw"
	case "sv", "swedish":
		return "sv"
	case "tl", "tagalog", "filipino", "tagalog(filipino)":
		return "tl"
	case "tg", "tajik":
		return "tg"
	case "ta", "tamil":
		return "ta"
	case "tt", "tatar":
		return "tt"
	case "te", "telugu":
		return "te"
	case "th", "thai":
		return "th"
	case "tr", "turkish":
		return "tr"
	case "tk", "turkmen":
		return "tk"
	case "uk", "ukrainian":
		return "uk"
	case "ur", "urdu":
		return "ur"
	case "ug", "uyghur":
		return "ug"
	case "uz", "uzbek":
		return "uz"
	case "vi", "vietnamese":
		return "vi"
	case "cy", "welsh":
		return "cy"
	case "xh", "xhosa":
		return "xh"
	case "yi", "yiddish":
		return "yi"
	case "yo", "yoruba":
		return "yo"
	case "zu", "zulu":
		return "zu"
	}
	return ""
}

func getLanguageName(language string) (langName string) {
	language = strings.ToLower(language)

	switch language {
	case "auto":
		return "(Automatically Detected)"
	case "af", "afrikaans":
		return "Afrikaans"
	case "sq", "albanian":
		return "Albanian"
	case "am", "amharic":
		return "Amharic"
	case "ar", "arabic":
		return "Arabic"
	case "hy", "armenian":
		return "Armenian"
	case "az", "azerbaijani":
		return "Azerbaijani"
	case "eu", "basque":
		return "Basque"
	case "be", "belarusian":
		return "Belarusian"
	case "bn", "bengali":
		return "Bengali"
	case "bs", "bosnian":
		return "Bosnian"
	case "bg", "bulgarian":
		return "bulgarian"
	case "ca", "catalan":
		return "Catalan"
	case "ceb", "cebuano":
		return "Cebuano"
	case "zh", "zh-CN", "chinese", "chinesesimpl", "chinesesimplified", "chinese(simplified)":
		return "Chinese (Simplified)"
	case "zh-tw", "chinesetrad", "chinesetradition", "chinesetraditional", "chinese(traditional)":
		return "Chinese (Traditional)"
	case "co", "corsican":
		return "Corsican"
	case "hr", "croatian":
		return "Croatian"
	case "cs", "czech":
		return "Czech"
	case "da", "danish":
		return "Danish"
	case "nl", "dutch":
		return "Dutch"
	case "en", "en-US", "en-CA", "en-UK", "en-AU", "english":
		return "English"
	case "eo", "esperanto":
		return "Esperanto"
	case "et", "estonian":
		return "Estonian"
	case "fi", "finnish":
		return "Finnish"
	case "fr", "french":
		return "French"
	case "fy", "frisian":
		return "Frisian"
	case "gl", "galician":
		return "Galician"
	case "ka", "georgian":
		return "Georgian"
	case "de", "german":
		return "German"
	case "el", "greek":
		return "Greek"
	case "gu", "gujarati":
		return "Gujarati"
	case "ht", "haitian", "haitiancreole", "creole":
		return "Haitian Creole"
	case "ha", "hausa":
		return "Hausa"
	case "haw", "hawaiian":
		return "Hawaiian"
	case "he", "iw", "hebrew":
		return "Hebrew"
	case "hi", "hindi":
		return "Hindi"
	case "hmn", "hmong":
		return "Hmong"
	case "hu", "hungarian":
		return "Hungarian"
	case "is", "icelandic":
		return "Icelandic"
	case "ig", "igbo":
		return "Igbo"
	case "id", "indonesian":
		return "Indonesian"
	case "ga", "irish":
		return "Irish"
	case "it", "italian":
		return "Italian"
	case "ja", "japanese", "weeaboo", "weeb":
		return "Japanese"
	case "jv", "javanese":
		return "Javanese"
	case "kn", "kannada":
		return "Kannada"
	case "kk", "kazakh":
		return "Kazakh"
	case "km", "khmer":
		return "Khmer"
	case "rw", "kinyarwanda":
		return "Kinyarwanda"
	case "ko", "korean":
		return "Korean"
	case "ku", "kurdish":
		return "Kurdish"
	case "ky", "kyrgyz":
		return "Kyrgyz"
	case "lo", "lao":
		return "Lao"
	case "la", "latin":
		return "Latin"
	case "lv", "latvian":
		return "Latvian"
	case "lt", "lithuanian":
		return "Lithuanian"
	case "lb", "luxembourgish":
		return "Luxembourgish"
	case "mk", "macedonian":
		return "Macedonian"
	case "mg", "malagasy":
		return "Malagasy"
	case "ms", "malay":
		return "Malay"
	case "ml", "malayalam":
		return "Malayalam"
	case "mt", "maltese":
		return "Maltese"
	case "mi", "maori":
		return "Maori"
	case "mr", "marathi":
		return "Marathi"
	case "mn", "mongolian":
		return "Mongolian"
	case "my", "myanmar", "burmese", "myanmarburmese", "myanmar(burmese)":
		return "Myanmar (Burmese)"
	case "ne", "nepali":
		return "Nepali"
	case "no", "norwegian":
		return "Norwegian"
	case "ny", "nyanja", "chichewa", "nyanjachichewa", "nyanja(chichewa)":
		return "Nyanja (Chichewa)"
	case "or", "odia", "oriya", "odiaoriya", "odia(oriya)":
		return "Odia (Oriya)"
	case "ps", "pashto":
		return "Pashto"
	case "fa", "persian":
		return "Persian"
	case "pl", "polish":
		return "Polish"
	case "pt", "portuguese", "portugal", "brazil":
		return "Portuguese"
	case "pa", "punjabi":
		return "Punjabi"
	case "ro", "romanian":
		return "Romanian"
	case "ru", "russian":
		return "Russian"
	case "sm", "samoan":
		return "Samoan"
	case "gd", "scots", "gaelic", "scotsgaelic":
		return "Scots Gaelic"
	case "sr", "serbian":
		return "Serbian"
	case "st", "sesotho":
		return "Sesotho"
	case "sn", "shona":
		return "Shona"
	case "sd", "sindhi":
		return "Sindhi"
	case "si", "sinhala", "sinhalese", "sinhalasinhalese", "sinhala(sinhalese)":
		return "Sinhala (Sinhalese)"
	case "sk", "slovak":
		return "Slovak"
	case "sl", "slovenian":
		return "Slovenian"
	case "so", "somali":
		return "Somali"
	case "es", "spanish":
		return "Spanish"
	case "su", "sundanese":
		return "Sundanese"
	case "sw", "swahili":
		return "Swahili"
	case "sv", "swedish":
		return "Swedish"
	case "tl", "tagalog", "filipino", "tagalog(filipino)":
		return "Tagalog (Filipino)"
	case "tg", "tajik":
		return "Tajik"
	case "ta", "tamil":
		return "Tamil"
	case "tt", "tatar":
		return "Tatar"
	case "te", "telugu":
		return "Telugu"
	case "th", "thai":
		return "Thai"
	case "tr", "turkish":
		return "Turkish"
	case "tk", "turkmen":
		return "Turkmen"
	case "uk", "ukrainian":
		return "Ukrainian"
	case "ur", "urdu":
		return "Urdu"
	case "ug", "uyghur":
		return "Uyghur"
	case "uz", "uzbek":
		return "Uzbek"
	case "vi", "vietnamese":
		return "Vietnamese"
	case "cy", "welsh":
		return "Welsh"
	case "xh", "xhosa":
		return "Xhosa"
	case "yi", "yiddish":
		return "Yiddish"
	case "yo", "yoruba":
		return "Yoruba"
	case "zu", "zulu":
		return "Zulu"
	}
	return ""
}
