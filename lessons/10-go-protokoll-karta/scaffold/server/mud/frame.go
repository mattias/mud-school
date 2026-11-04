package mud

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
	// TODO(lektion 10): "text" -> texten med \r\n, "map" -> raderna med \r\n,
	// allt annat -> "".
	return ""
}

// encodeJSON gör om en Frame till en rad JSON (avslutad med \r\n).
func encodeJSON(f Frame) string {
	// TODO(lektion 10): json.Marshal
	return ""
}
