package mud

import (
	"encoding/json"
	"log"
	"strings"
)

// HelloPrefix is what a graphical client sends to get JSON instead of text.
const HelloPrefix = "@hello"

// Frame is a message from the server to the client. See docs/<lang>/PROTOCOL.md.
//
// A telnet client only gets the text. A client that has sent "@hello" gets
// every Frame as one line of JSON:
//
//	{"type":"text","text":"You say: hello"}
//	{"type":"map","map":{"width":5,"height":5,"rows":[...],"legend":{...}}}
//	{"type":"status","status":{"name":"Anna","room":"Mudby Square",...}}
type Frame struct {
	Type   string     `json:"type"` // "hello", "text", "map", "status", "prompt", "zone" or "delta"
	Text   string     `json:"text,omitempty"`
	Kind   string     `json:"kind,omitempty"` // for text: one of the Kind constants
	Map    *MapData   `json:"map,omitempty"`
	Status *Status    `json:"status,omitempty"`
	Prompt string     `json:"prompt,omitempty"` // for prompt: what is asked for (lesson 20)
	Zone   *ZoneData  `json:"zone,omitempty"`   // for zone: the whole zone (lesson 23)
	Delta  *DeltaData `json:"delta,omitempty"`  // for delta: what changed (lesson 23)
}

// MapData is a piece of the map around the player. Every character in Rows is
// a code that Legend translates to a meaning ("town", "forest", "self", ...).
// The client decides how a code is drawn – as a coloured character, an image ...
// The server sends MEANING, the client decides APPEARANCE.
type MapData struct {
	Width  int               `json:"width"`
	Height int               `json:"height"`
	Rows   []string          `json:"rows"`
	Legend map[string]string `json:"legend"`
}

// Status is what the client shows in its status bar.
type Status struct {
	Name    string  `json:"name"`
	Room    string  `json:"room"`
	Terrain string  `json:"terrain"`
	HP      int     `json:"hp"`
	MaxHP   int     `json:"max_hp"`
	Online  int     `json:"online"`
	Target  *Target `json:"target,omitempty"` // who you are fighting (lesson 18)
}

// Target is the opponent shown in the client's status bar.
type Target struct {
	Name  string `json:"name"`
	HP    int    `json:"hp"`
	MaxHP int    `json:"max_hp"`
}

// What kind of text a text frame is, so that a client can colour it.
// Clients must treat an unknown kind like KindInfo (see PROTOCOL.md).
const (
	KindInfo   = "info"   // everything else (the default)
	KindRoom   = "room"   // look: a room description
	KindSay    = "say"    // someone talks: say, shout
	KindMove   = "move"   // someone comes or goes
	KindError  = "error"  // the command didn't work
	KindCombat = "combat" // fights (lesson 18)
)

// TextFrame is a Frame with ordinary text.
func TextFrame(text string) Frame {
	return Frame{Type: "text", Text: text}
}

// KindFrame is a text frame of a certain kind.
func KindFrame(kind, text string) Frame {
	return Frame{Type: "text", Text: text, Kind: kind}
}

// encodeText turns a Frame into what a telnet client should see.
// An empty string means "send nothing".
func encodeText(f Frame) string {
	switch f.Type {
	case "text":
		return strings.ReplaceAll(f.Text, "\n", "\r\n") + "\r\n"
	case "map":
		return strings.Join(f.Map.Rows, "\r\n") + "\r\n"
	case "zone":
		return renderZone(f.Zone)
	default:
		return "" // status bars etc. are only for graphical clients
	}
}

// encodeJSON turns a Frame into one line of JSON (ending in \r\n).
func encodeJSON(f Frame) string {
	data, err := json.Marshal(f)
	if err != nil {
		log.Printf("cannot encode %+v: %v", f, err)
		return ""
	}
	return string(data) + "\r\n"
}
