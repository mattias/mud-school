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
	var plain string
	if err := json.Unmarshal(data, &plain); err == nil {
		*t = Text{"": plain}
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return errors.New(MsgErrText)
	}
	*t = m
	return nil
}

// MarshalJSON writes a text with only a plain string back as a string, so
// that saving a world (lesson 36) keeps the file as it was.
func (t Text) MarshalJSON() ([]byte, error) {
	if s, ok := t[""]; ok && len(t) == 1 {
		return json.Marshal(s)
	}
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
	for _, l := range []string{lang, fallback, ""} {
		if s, ok := t[l]; ok {
			return s
		}
	}
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return t[keys[0]]
}

// TextData is what a text's template can use: {{.Name}}, {{.Step}}, {{.Steps}}.
type TextData struct {
	Name        string // the player's name
	Step, Steps int    // the tutorial step, and how many there are
}

// Render executes the text in lang as a text/template with data. A text
// that is not a valid template is shown as it is.
func (t Text) Render(lang, fallback string, data TextData) string {
	s := t.In(lang, fallback)
	tmpl, err := template.New("text").Option("missingkey=error").Parse(s)
	if err != nil {
		return s
	}
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return s
	}
	return b.String()
}

// check reports the first language whose text is not a valid template, or
// uses a field TextData doesn't have.
func (t Text) check() error {
	for _, lang := range sortedKeys(t) {
		tmpl, err := template.New("text").Option("missingkey=error").Parse(t[lang])
		if err == nil {
			err = tmpl.Execute(new(bytes.Buffer), TextData{})
		}
		if err != nil {
			return fmt.Errorf(MsgErrTemplate, lang, err)
		}
	}
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
	set := map[string]bool{}
	if w.Language != "" {
		set[w.Language] = true
	}
	add := func(t Text) {
		for k := range t {
			if k != "" {
				set[k] = true
			}
		}
	}
	for _, r := range w.Rooms {
		for k := range r.Translations {
			set[k] = true
		}
	}
	if t := w.Tutorial; t != nil {
		add(t.Done)
		for _, st := range t.Steps {
			add(st.Text)
		}
	}
	return sortedKeys(set)
}

// roomText is the room's name and description in the player's language.
func (s *Server) roomText(p *Player, r *Room) RoomText {
	for _, l := range []string{p.lang, s.world.Language} {
		if tr, ok := r.Translations[l]; ok && l != "" {
			return tr
		}
	}
	return RoomText{Name: r.Name, Description: r.Description}
}

// render renders a world text for the player, in their language.
func (s *Server) render(p *Player, t Text, data TextData) string {
	data.Name = p.Name
	return t.Render(p.lang, s.world.Language, data)
}

// cmdLanguage: "language" shows the player's language and the ones there
// are; "language sv" chooses one.
func cmdLanguage(s *Server, p *Player, cmd Command) {
	langs := s.world.Languages()
	if cmd.Rest == "" {
		current := p.lang
		if current == "" {
			current = s.world.Language
		}
		p.SendKind(KindInfo, fmt.Sprintf(MsgLanguage, current, strings.Join(langs, MsgListSep)))
		return
	}
	lang := strings.ToLower(cmd.Rest)
	if !slices.Contains(langs, lang) {
		p.SendKind(KindError, fmt.Sprintf(MsgLanguageUnknown, lang, strings.Join(langs, MsgListSep)))
		return
	}
	p.lang = lang
	p.SendKind(KindInfo, fmt.Sprintf(MsgLanguageSet, lang))
	s.look(p)
	s.showTutorialStep(p)
}
