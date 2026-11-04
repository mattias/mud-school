package mud

import (
	"encoding/json"
	"strings"
	"testing"
)

func mapTestWorld() *World {
	return NewWorld("torg",
		&Room{ID: "torg", Name: "Torget", Terrain: "town", X: 0, Y: 0, Exits: map[string]string{"north": "krog", "east": "smedja"}},
		&Room{ID: "krog", Name: "Krogen", Terrain: "inside", X: 0, Y: -1, Exits: map[string]string{"south": "torg", "up": "vind"}},
		&Room{ID: "vind", Name: "Vinden", Terrain: "inside", X: 0, Y: -1, Z: 1, Exits: map[string]string{"down": "krog"}},
		&Room{ID: "smedja", Name: "Smedjan", Terrain: "inside", X: 1, Y: 0, Exits: map[string]string{"west": "torg"}},
		&Room{ID: "ang", Name: "Ängen", Terrain: "field", X: -1, Y: 1},
		&Room{ID: "langtbort", Name: "Långt bort", Terrain: "water", X: 5, Y: 5},
	)
}

func TestRenderMap(t *testing.T) {
	w := mapTestWorld()
	m := w.RenderMap(w.Room("torg"), 1)
	want := []string{
		"  <  ",
		"  |  ",
		"  #-+",
		"     ",
		`"    `,
	}
	if m.Width != 5 || m.Height != 5 {
		t.Errorf("Width=%d Height=%d, väntade 5 och 5", m.Width, m.Height)
	}
	if strings.Join(m.Rows, "\n") != strings.Join(want, "\n") {
		t.Fatalf("fel karta.\nfick:\n%s\nväntade:\n%s", showRows(m.Rows), showRows(want))
	}
	for _, row := range m.Rows {
		for _, ch := range row {
			if ch != ' ' && m.Legend[string(ch)] == "" {
				t.Errorf("tecknet %q saknas i Legend %v", ch, m.Legend)
			}
		}
	}
	for code, name := range map[string]string{"#": "town", "+": "inside", `"`: "field", "<": "stairs_up", "|": "path_ns", "-": "path_ew"} {
		if m.Legend[code] != name {
			t.Errorf("Legend[%q] = %q, väntade %q", code, m.Legend[code], name)
		}
	}
}

func TestRenderMapOtherFloor(t *testing.T) {
	w := mapTestWorld()
	m := w.RenderMap(w.Room("vind"), 1)
	if got := strings.Join(m.Rows, ""); strings.TrimSpace(got) != ">" {
		t.Errorf("på vinden ska bara vinden synas (som '>'), fick:\n%s", showRows(m.Rows))
	}
}

func TestWithSelf(t *testing.T) {
	w := mapTestWorld()
	m := w.RenderMap(w.Room("torg"), 1).WithSelf()
	if m.Rows[2] != "  @-+" {
		t.Errorf("mittenraden ska vara %q, fick %q", "  @-+", m.Rows[2])
	}
	if m.Legend["@"] != "self" {
		t.Errorf(`Legend["@"] ska vara "self"`)
	}
	if orig := w.RenderMap(w.Room("torg"), 1); orig.Rows[2] != "  #-+" {
		t.Error("WithSelf får inte ändra originalkartan")
	}
}

func showRows(rows []string) string {
	return "  ┌" + strings.Repeat("─", len(rows[0])) + "┐\n  │" + strings.Join(rows, "│\n  │") + "│\n  └" + strings.Repeat("─", len(rows[0])) + "┘"
}

// readFrame läser en rad och avkodar den som JSON.
func (c *testClient) readFrame() map[string]any {
	c.t.Helper()
	line, ok := c.readLine(testTimeout)
	if !ok {
		c.t.Fatalf("%s: väntade en JSON-rad men servern stängde anslutningen", c.name)
	}
	var f map[string]any
	if err := json.Unmarshal([]byte(line), &f); err != nil {
		c.t.Fatalf("%s: raden %q är inte giltig JSON: %v", c.name, line, err)
	}
	if _, ok := f["type"].(string); !ok {
		c.t.Fatalf("%s: frame utan \"type\": %q", c.name, line)
	}
	return f
}

// expectFrame läser frames tills en av typen typ kommer.
func (c *testClient) expectFrame(typ string) map[string]any {
	c.t.Helper()
	for i := 0; i < 20; i++ {
		if f := c.readFrame(); f["type"] == typ {
			return f
		}
	}
	c.t.Fatalf("%s: fick ingen frame av typen %q", c.name, typ)
	return nil
}

func jsonLogin(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expect(Welcome)
	c.expect("Vad heter du?")
	c.send("@hello mudschool 1")
	if f := c.readFrame(); f["type"] != "hello" {
		t.Fatalf(`svaret på @hello ska vara {"type":"hello",...}, fick %v`, f)
	}
	c.send(name)
	f := c.expectFrame("text")
	if f["text"] != "Välkommen, "+name+"!" {
		t.Fatalf("väntade text-frame %q, fick %v", "Välkommen, "+name+"!", f)
	}
	return c
}

func TestJSONClientGetsTextMapAndStatus(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(mapTestWorld()))
	c := jsonLogin(t, addr, "Anna")

	look := c.expectFrame("text")
	text, _ := look["text"].(string)
	if !strings.HasPrefix(text, "Torget\n") || strings.Contains(text, "\r") {
		t.Errorf("look ska komma som EN text-frame med \\n mellan raderna, fick %q", text)
	}

	mf := c.expectFrame("map")
	m, ok := mf["map"].(map[string]any)
	if !ok {
		t.Fatalf(`map-frame saknar "map": %v`, mf)
	}
	rows, _ := m["rows"].([]any)
	if len(rows) == 0 || m["width"] == nil || m["height"] == nil || m["legend"] == nil {
		t.Fatalf("map ska ha width, height, rows och legend: %v", m)
	}
	mid, _ := rows[len(rows)/2].(string)
	if len(mid) == 0 || mid[len(mid)/2] != '@' {
		t.Errorf("spelaren '@' ska stå mitt på kartan, mittenraden är %q", mid)
	}

	sf := c.expectFrame("status")
	st, _ := sf["status"].(map[string]any)
	if st["name"] != "Anna" || st["room"] != "Torget" || st["terrain"] != "town" {
		t.Errorf("status ska ha name, room och terrain, fick %v", st)
	}
	for _, key := range []string{"hp", "max_hp", "online"} {
		if _, ok := st[key].(float64); !ok {
			t.Errorf("status saknar talet %q: %v", key, st)
		}
	}

	c.send("n")
	if f := c.expectFrame("status"); f["status"].(map[string]any)["room"] != "Krogen" {
		t.Errorf("efter north ska status visa Krogen, fick %v", f)
	}
}

func TestJSONAndTextClientsTogether(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(mapTestWorld()))
	anna := jsonLogin(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	bertil.drain()

	bertil.send("say hej Anna")
	for {
		f := anna.expectFrame("text")
		if f["text"] == "Bertil säger: hej Anna" {
			break
		}
	}
	anna.send("say hej Bertil")
	bertil.expectContains("Anna säger: hej Bertil")

	bertil.send("look")
	for i := 0; i < 5; i++ {
		line, _ := bertil.readLine(testTimeout)
		if strings.HasPrefix(line, "{") {
			t.Fatalf("en telnet-klient ska aldrig få JSON, fick %q", line)
		}
		if strings.HasPrefix(line, "Här finns") {
			break
		}
	}
}

func TestTextClientMapCommand(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(mapTestWorld()))
	c := login(t, addr, "Anna")
	c.drain()
	c.send("map")
	c.expectContains("@-+")
}
