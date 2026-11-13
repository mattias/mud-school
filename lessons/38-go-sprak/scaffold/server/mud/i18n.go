package mud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"text/template"
)

// Languages per player (lesson 38): the world's texts – the tutorial, and
// rooms that have translations – are shown in the language each player
// chooses with "language". The server's own messages (messages.go) stay in
// the language the server was built in.

// Text is a text in one or more languages. In a world file it is either a
// plain string or an object with one string per language:
//
//	"text": "Type look."
//	"text": {"en": "Type look.", "sv": "Skriv look."}
//
// A plain string is stored under the key "".
type Text map[string]string

// UnmarshalJSON makes Text a json.Unmarshaler: encoding/json calls it for
// every Text it decodes.
func (t *Text) UnmarshalJSON(data []byte) error {
	// TODO(lesson 38): first try a string ({"": it}), then a map[string]string;
	// neither: errors.New(MsgErrText).
	_, _ = json.Unmarshal, errors.New
	return errors.New("TODO: Text.UnmarshalJSON")
}

// MarshalJSON writes a text with only a plain string back as a string, so
// that saving a world (lesson 36) keeps the file as it was.
func (t Text) MarshalJSON() ([]byte, error) {
	// TODO(lesson 38): only "": a string; otherwise the map.
	return json.Marshal(map[string]string(t))
}

// empty reports whether the text has nothing to say in any language –
// "", {"en": ""} and null are all empty.
func (t Text) empty() bool {
	for _, s := range t {
		if s != "" {
			return false
		}
	}
	return true
}

// In returns the text in lang. If there is none it falls back to the
// world's default language, then the plain string, then the first language
// in alphabetical order – something is always better than nothing.
func (t Text) In(lang, fallback string) string {
	// TODO(lesson 38): lang, fallback, "" – then the first key in sorted order.
	return t[""]
}

// TextData is what a text's template can use: {{.Name}}, {{.Step}}, {{.Steps}}.
type TextData struct {
	Name        string // the player's name
	Step, Steps int    // the tutorial step, and how many there are
}

// Render executes the text in lang as a text/template with data. A text
// that is not a valid template is shown as it is.
func (t Text) Render(lang, fallback string, data TextData) string {
	// TODO(lesson 38): t.In(...) run as a text/template (Option
	// "missingkey=error") with data; on any error the text as it is.
	_ = template.New
	_ = bytes.Buffer{}
	return t.In(lang, fallback)
}

// check reports the first language whose text is not a valid template, or
// uses a field TextData doesn't have.
func (t Text) check() error {
	// TODO(lesson 38): parse and execute every language with TextData{};
	// the first that fails: fmt.Errorf(MsgErrTemplate, lang, err).
	_ = fmt.Errorf
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// RoomText is a room's name and description in one language.
type RoomText struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Languages lists every language the world has texts in, sorted.
func (w *World) Languages() []string {
	// TODO(lesson 38): w.Language, the room translations' keys, and the keys
	// (not "") of the tutorial's texts – sorted, each once.
	return nil
}

// roomText is the room's name and description in the player's language.
func (s *Server) roomText(p *Player, r *Room) RoomText {
	// TODO(lesson 38): the translation for p.lang, else for the world's
	// language, else the room's own name and description.
	return RoomText{Name: r.Name, Description: r.Description}
}

// render renders a world text for the player, in their language.
func (s *Server) render(p *Player, t Text, data TextData) string {
	// TODO(lesson 38): data.Name = p.Name, and Render in p.lang with the
	// world's language as fallback.
	return t.In("", "")
}

// cmdLanguage: "language" shows the player's language and the ones there
// are; "language sv" chooses one.
func cmdLanguage(s *Server, p *Player, cmd Command) {
	// TODO(lesson 38): no argument: MsgLanguage (the player's language, or the
	// world's) with the list; an unknown one: MsgLanguageUnknown; otherwise set
	// p.lang, MsgLanguageSet, look and the tutorial step again.
	_, _ = strings.Join, slices.Contains[[]string]
}
