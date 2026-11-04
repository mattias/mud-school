package mud

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
	Type   string   `json:"type"` // "hello", "text", "map" or "status"
	Text   string   `json:"text,omitempty"`
	Map    *MapData `json:"map,omitempty"`
	Status *Status  `json:"status,omitempty"`
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
	Name    string `json:"name"`
	Room    string `json:"room"`
	Terrain string `json:"terrain"`
	HP      int    `json:"hp"`
	MaxHP   int    `json:"max_hp"`
	Online  int    `json:"online"`
}

// TextFrame is a Frame with ordinary text.
func TextFrame(text string) Frame {
	return Frame{Type: "text", Text: text}
}

// encodeText turns a Frame into what a telnet client should see.
// An empty string means "send nothing".
func encodeText(f Frame) string {
	// TODO(lesson 10): "text" -> the text with \r\n, "map" -> the rows with \r\n,
	// everything else -> "".
	return ""
}

// encodeJSON turns a Frame into one line of JSON (ending in \r\n).
func encodeJSON(f Frame) string {
	// TODO(lesson 10): json.Marshal
	return ""
}
