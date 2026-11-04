package main

import (
	"os"
	"runtime"
	"strings"
)

// Languages lists the supported course languages.
var Languages = []string{"en", "sv", "ja"}

// lang is the active language for this run of the tool.
var lang = "en"

// self is how the user starts the tool: "./mudschool" with a POSIX shell,
// ".\mudschool.cmd" on Windows (cmd.exe and PowerShell).
var self = func() string {
	if runtime.GOOS == "windows" {
		return `.\mudschool.cmd`
	}
	return "./mudschool"
}()

// fixSelf rewrites "./mudschool" in text shown to the user so that the
// commands can be copied as they are on every platform.
func fixSelf(s string) string {
	if self == "./mudschool" {
		return s
	}
	return strings.ReplaceAll(s, "./mudschool", self)
}

// L picks the text for the active language.
func L(sv, en, ja string) string {
	switch lang {
	case "sv":
		return fixSelf(sv)
	case "ja":
		return fixSelf(ja)
	default:
		return fixSelf(en)
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

// detectLang guesses a language from the environment. $MUDSCHOOL_LANG wins
// if it names a supported language. Otherwise the POSIX rule applies: the
// first non-empty of $LC_ALL, $LC_MESSAGES and $LANG decides (so LC_ALL=C
// means English even if LANG=sv_SE.UTF-8). Falls back to English.
func detectLang() string {
	if l := langFromLocale(os.Getenv("MUDSCHOOL_LANG")); l != "" {
		return l
	}
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := os.Getenv(v); val != "" {
			if l := langFromLocale(val); l != "" {
				return l
			}
			return "en"
		}
	}
	return "en"
}

// langFromLocale maps a locale like "sv_SE.UTF-8" or "ja" to a course
// language, or "" if it is not one of ours.
func langFromLocale(val string) string {
	val = strings.ToLower(val)
	for _, l := range Languages {
		if val == l || strings.HasPrefix(val, l+"_") || strings.HasPrefix(val, l+"-") || strings.HasPrefix(val, l+".") || strings.HasPrefix(val, l+"@") {
			return l
		}
	}
	return ""
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
