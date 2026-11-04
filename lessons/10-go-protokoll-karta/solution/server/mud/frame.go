package mud

import (
	"encoding/json"
	"log"
	"strings"
)

// HelloPrefix är det en grafisk klient skickar för att få JSON istället för text.
const HelloPrefix = "@hello"

// Frame är ett meddelande från servern till klienten. Se docs/PROTOKOLL.md.
//
// En telnet-klient får bara texten. En klient som har skickat "@hello" får
// varje Frame som en rad JSON:
//
//	{"type":"text","text":"Du säger: hej"}
//	{"type":"map","map":{"width":5,"height":5,"rows":[...],"legend":{...}}}
//	{"type":"status","status":{"name":"Anna","room":"Torget",...}}
type Frame struct {
	Type   string   `json:"type"` // "hello", "text", "map" eller "status"
	Text   string   `json:"text,omitempty"`
	Map    *MapData `json:"map,omitempty"`
	Status *Status  `json:"status,omitempty"`
}

// MapData är en bit av kartan runt spelaren. Varje tecken i Rows är en kod
// som Legend översätter till ett namn ("town", "forest", "self", ...).
// Klienten bestämmer själv hur en kod ska ritas – som ett färgat tecken, en
// bild ... Servern skickar BETYDELSE, klienten bestämmer UTSEENDE.
type MapData struct {
	Width  int               `json:"width"`
	Height int               `json:"height"`
	Rows   []string          `json:"rows"`
	Legend map[string]string `json:"legend"`
}

// Status är det som visas i klientens statusrad.
type Status struct {
	Name    string `json:"name"`
	Room    string `json:"room"`
	Terrain string `json:"terrain"`
	HP      int    `json:"hp"`
	MaxHP   int    `json:"max_hp"`
	Online  int    `json:"online"`
}

// TextFrame är en Frame med vanlig text.
func TextFrame(text string) Frame {
	return Frame{Type: "text", Text: text}
}

// encodeText gör om en Frame till det en telnet-klient ska se.
// Tom sträng betyder "skicka ingenting".
func encodeText(f Frame) string {
	switch f.Type {
	case "text":
		return strings.ReplaceAll(f.Text, "\n", "\r\n") + "\r\n"
	case "map":
		return strings.Join(f.Map.Rows, "\r\n") + "\r\n"
	default:
		return "" // statusrader m.m. är bara för grafiska klienter
	}
}

// encodeJSON gör om en Frame till en rad JSON.
func encodeJSON(f Frame) string {
	data, err := json.Marshal(f)
	if err != nil {
		log.Printf("kunde inte koda %+v: %v", f, err)
		return ""
	}
	return string(data) + "\r\n"
}
