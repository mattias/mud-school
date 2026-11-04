package main

import (
	"os"
	"strings"
)

// Languages lists the supported course languages.
var Languages = []string{"sv", "en", "ja"}

// lang is the active language for this run of the tool.
var lang = "en"

// L picks the text for the active language.
func L(sv, en, ja string) string {
	switch lang {
	case "sv":
		return sv
	case "ja":
		return ja
	default:
		return en
	}
}

func validLang(l string) bool {
	for _, x := range Languages {
		if x == l {
			return true
		}
	}
	return false
}

// detectLang guesses a language from the environment ($MUDSCHOOL_LANG,
// $LC_ALL, $LANG). Falls back to English.
func detectLang() string {
	for _, v := range []string{"MUDSCHOOL_LANG", "LC_ALL", "LC_MESSAGES", "LANG"} {
		val := strings.ToLower(os.Getenv(v))
		for _, l := range Languages {
			if strings.HasPrefix(val, l) {
				return l
			}
		}
	}
	return "en"
}

// extractLangFlag removes "--lang xx" / "--lang=xx" from args.
func extractLangFlag(args []string) (rest []string, value string) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--lang" && i+1 < len(args):
			value = args[i+1]
			i++
		case strings.HasPrefix(a, "--lang="):
			value = strings.TrimPrefix(a, "--lang=")
		default:
			rest = append(rest, a)
		}
	}
	return rest, value
}
