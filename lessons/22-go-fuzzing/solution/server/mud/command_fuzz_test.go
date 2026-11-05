package mud

import (
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// FuzzParseCommand feeds ParseCommand random lines and checks what must ALWAYS
// be true, whatever the line is. Run it with:
//
//	go test -fuzz=FuzzParseCommand -fuzztime=30s ./mud/
//
// A plain `go test` runs only the seeds below (and anything saved in
// testdata/fuzz/FuzzParseCommand/), so it doubles as an ordinary test.
func FuzzParseCommand(f *testing.F) {
	for _, seed := range []string{
		"", "look", "n", "Get Sword from chest", "'hello there", "  say   hi  ",
		"look\tsword", "ÅÄÖ", "say  hi",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, line string) {
		c := ParseCommand(line)

		if strings.IndexFunc(c.Verb, unicode.IsSpace) >= 0 {
			t.Errorf("the verb %q contains whitespace", c.Verb)
		}
		if c.Rest != strings.TrimSpace(c.Rest) {
			t.Errorf("Rest %q is not trimmed", c.Rest)
		}
		if !slices.Equal(c.Args, strings.Fields(c.Rest)) {
			t.Errorf("Args %q are not the words of Rest %q", c.Args, c.Rest)
		}
		if (c.Verb == "") != (strings.TrimSpace(line) == "") {
			t.Errorf("ParseCommand(%q).Verb = %q", line, c.Verb)
		}
		if utf8.ValidString(line) && strings.ToLower(c.Verb) != c.Verb {
			t.Errorf("the verb %q is not lower case", c.Verb)
		}
	})
}
