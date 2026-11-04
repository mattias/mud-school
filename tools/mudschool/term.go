package main

import (
	"os"
	"regexp"
	"strings"
)

var useColor = os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"

func ansi(code, s string) string {
	s = fixSelf(s)
	if !useColor {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func red(s string) string    { return ansi("31", s) }
func green(s string) string  { return ansi("32", s) }
func yellow(s string) string { return ansi("33", s) }
func cyan(s string) string   { return ansi("36", s) }
func dim(s string) string    { return ansi("2", s) }
func bold(s string) string   { return ansi("1", s) }

var (
	inlineCode = regexp.MustCompile("`([^`]+)`")
	boldText   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
)

// renderMarkdown gör en enkel terminalvänlig rendering av Markdown.
func renderMarkdown(md string) string {
	var b strings.Builder
	inCode := false
	for _, line := range strings.Split(md, "\n") {
		switch {
		case strings.HasPrefix(line, "```"):
			inCode = !inCode
			b.WriteString(dim("  ────") + "\n")
			continue
		case inCode:
			b.WriteString("  " + yellow(line) + "\n")
			continue
		case strings.HasPrefix(line, "# "):
			b.WriteString(bold(cyan(strings.ToUpper(line[2:]))) + "\n")
			continue
		case strings.HasPrefix(line, "## "), strings.HasPrefix(line, "### "):
			b.WriteString(bold(cyan(strings.TrimLeft(line, "# "))) + "\n")
			continue
		case strings.HasPrefix(line, "> "):
			line = dim("│ ") + line[2:]
		}
		line = inlineCode.ReplaceAllString(line, yellow("$1"))
		line = boldText.ReplaceAllString(line, bold("$1"))
		b.WriteString(line + "\n")
	}
	return fixSelf(b.String())
}
