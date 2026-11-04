package mud

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func mapTestWorld() *World {
	return NewWorld("torg",
		&Room{ID: "torg", Name: "Square", Terrain: "town", X: 0, Y: 0, Exits: map[string]string{"north": "krog", "east": "smedja"}},
		&Room{ID: "krog", Name: "Tavern", Terrain: "inside", X: 0, Y: -1, Exits: map[string]string{"south": "torg", "up": "vind"}},
		&Room{ID: "vind", Name: "Attic", Terrain: "inside", X: 0, Y: -1, Z: 1, Exits: map[string]string{"down": "krog"}},
		&Room{ID: "smedja", Name: "Smithy", Terrain: "inside", X: 1, Y: 0, Exits: map[string]string{"west": "torg"}},
		&Room{ID: "ang", Name: "Meadow", Terrain: "field", X: -1, Y: 1},
		&Room{ID: "langtbort", Name: "Far away", Terrain: "water", X: 5, Y: 5},
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
		t.Errorf(L("Width=%d Height=%d, väntade 5 och 5", "Width=%d Height=%d, expected 5 and 5", "Width=%d Height=%d ですが、5 と 5 を期待しました"), m.Width, m.Height)
	}
	if len(m.Rows) != len(want) || strings.Join(m.Rows, "\n") != strings.Join(want, "\n") {
		t.Fatalf(L("fel karta.\nfick:\n%s\nväntade:\n%s", "wrong map.\ngot:\n%s\nexpected:\n%s", "マップが違います。\n実際:\n%s\n期待:\n%s"), showRows(m.Rows), showRows(want))
	}
	for _, row := range m.Rows {
		for _, ch := range row {
			if ch != ' ' && m.Legend[string(ch)] == "" {
				t.Errorf(L("tecknet %q saknas i Legend %v", "the character %q is missing from Legend %v", "文字 %q が Legend %v にありません"), ch, m.Legend)
			}
		}
	}
	for code, name := range map[string]string{"#": "town", "+": "inside", `"`: "field", "<": "stairs_up", "|": "path_ns", "-": "path_ew"} {
		if m.Legend[code] != name {
			t.Errorf("Legend[%q] = %q, %s %q", code, m.Legend[code], L("väntade", "expected", "期待値"), name)
		}
	}
}

func TestRenderMapOtherFloor(t *testing.T) {
	w := mapTestWorld()
	m := w.RenderMap(w.Room("vind"), 1)
	if got := strings.Join(m.Rows, ""); strings.TrimSpace(got) != ">" {
		t.Errorf(L("på vinden ska bara vinden synas (som '>'), fick:\n%s",
			"in the attic only the attic should be visible (as '>'), got:\n%s",
			"屋根裏では屋根裏だけ（'>' として）表示されるべきですが、実際は:\n%s"), showRows(m.Rows))
	}
}

func TestWithSelf(t *testing.T) {
	w := mapTestWorld()
	m := w.RenderMap(w.Room("torg"), 1).WithSelf()
	if len(m.Rows) != 5 || m.Rows[2] != "  @-+" {
		t.Fatalf(L("mittenraden ska vara %q, fick %q", "the middle row should be %q, got %q", "中央の行は %q であるべきですが、%q でした"), "  @-+", m.Rows)
	}
	if m.Legend["@"] != "self" {
		t.Errorf(`Legend["@"] != "self"`)
	}
	if orig := w.RenderMap(w.Room("torg"), 1); orig.Rows[2] != "  #-+" {
		t.Error(L("WithSelf får inte ändra originalkartan", "WithSelf must not change the original map", "WithSelf は元のマップを変更してはいけません"))
	}
}

func showRows(rows []string) string {
	if len(rows) == 0 {
		return "  (" + L("tom", "empty", "空") + ")"
	}
	return "  ┌" + strings.Repeat("─", len(rows[0])) + "┐\n  │" + strings.Join(rows, "│\n  │") + "│\n  └" + strings.Repeat("─", len(rows[0])) + "┘"
}

// readFrame reads a line and decodes it as JSON.
func (c *testClient) readFrame() map[string]any {
	c.t.Helper()
	line, ok := c.readLine(testTimeout)
	if !ok {
		c.t.Fatalf(L("%s: väntade en JSON-rad men servern stängde anslutningen", "%s: expected a JSON line but the server closed the connection", "%s: JSON の行を期待しましたが、サーバーが接続を閉じました"), c.name)
	}
	var f map[string]any
	if err := json.Unmarshal([]byte(line), &f); err != nil {
		c.t.Fatalf(L("%s: raden %q är inte giltig JSON: %v", "%s: the line %q is not valid JSON: %v", "%s: 行 %q は正しい JSON ではありません: %v"), c.name, line, err)
	}
	if _, ok := f["type"].(string); !ok {
		c.t.Fatalf(L("%s: frame utan \"type\": %q", "%s: frame without \"type\": %q", "%s: \"type\" のないフレーム: %q"), c.name, line)
	}
	return f
}

// expectFrame reads frames until one of type typ arrives.
func (c *testClient) expectFrame(typ string) map[string]any {
	c.t.Helper()
	for i := 0; i < 20; i++ {
		if f := c.readFrame(); f["type"] == typ {
			return f
		}
	}
	c.t.Fatalf(L("%s: fick ingen frame av typen %q", "%s: got no frame of type %q", "%s: 種類 %q のフレームが届きませんでした"), c.name, typ)
	return nil
}

// expectText reads text frames until one with exactly want arrives.
func (c *testClient) expectText(want string) {
	c.t.Helper()
	for i := 0; i < 20; i++ {
		if f := c.expectFrame("text"); f["text"] == want {
			return
		}
	}
	c.t.Fatalf(L("%s: ingen text-frame med %q", "%s: no text frame with %q", "%s: %q のテキストフレームがありません"), c.name, want)
}

func jsonLogin(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expect(Welcome)
	c.expect(MsgAskName)
	c.send("@hello mudschool 1")
	if f := c.readFrame(); f["type"] != "hello" {
		t.Fatalf(L(`svaret på @hello ska vara {"type":"hello",...}, fick %v`, `the reply to @hello should be {"type":"hello",...}, got %v`, `@hello への応答は {"type":"hello",...} であるべきですが、%v でした`), f)
	}
	c.send(name)
	welcome := fmt.Sprintf(MsgWelcomeName, name)
	if f := c.expectFrame("text"); f["text"] != welcome {
		t.Fatalf(L("väntade text-frame %q, fick %v", "expected the text frame %q, got %v", "テキストフレーム %q を期待しましたが、%v でした"), welcome, f)
	}
	return c
}

func TestJSONClientGetsTextMapAndStatus(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(mapTestWorld()))
	c := jsonLogin(t, addr, "Anna")

	look := c.expectFrame("text")
	text, _ := look["text"].(string)
	if !strings.HasPrefix(text, "Square\n") || strings.Contains(text, "\r") {
		t.Errorf(L("look ska komma som EN text-frame med \\n mellan raderna, fick %q",
			"look should arrive as ONE text frame with \\n between the lines, got %q",
			"look は行の間が \\n の 1 つのテキストフレームで届くべきですが、%q でした"), text)
	}

	mf := c.expectFrame("map")
	m, ok := mf["map"].(map[string]any)
	if !ok {
		t.Fatalf(L(`map-frame saknar "map": %v`, `map frame without "map": %v`, `map フレームに "map" がありません: %v`), mf)
	}
	rows, _ := m["rows"].([]any)
	if len(rows) == 0 || m["width"] == nil || m["height"] == nil || m["legend"] == nil {
		t.Fatalf(L("map ska ha width, height, rows och legend: %v", "map should have width, height, rows and legend: %v", "map には width, height, rows, legend が必要です: %v"), m)
	}
	mid, _ := rows[len(rows)/2].(string)
	if len(mid) == 0 || mid[len(mid)/2] != '@' {
		t.Errorf(L("spelaren '@' ska stå mitt på kartan, mittenraden är %q",
			"the player '@' should be in the middle of the map, the middle row is %q",
			"プレイヤー '@' はマップの中央にいるべきですが、中央の行は %q です"), mid)
	}

	sf := c.expectFrame("status")
	st, _ := sf["status"].(map[string]any)
	if st["name"] != "Anna" || st["room"] != "Square" || st["terrain"] != "town" {
		t.Errorf(L("status ska ha name, room och terrain, fick %v", "status should have name, room and terrain, got %v", "status には name, room, terrain が必要ですが、%v でした"), st)
	}
	for _, key := range []string{"hp", "max_hp", "online"} {
		if _, ok := st[key].(float64); !ok {
			t.Errorf(L("status saknar talet %q: %v", "status is missing the number %q: %v", "status に数値 %q がありません: %v"), key, st)
		}
	}

	c.send("n")
	if f := c.expectFrame("status"); f["status"].(map[string]any)["room"] != "Tavern" {
		t.Errorf(L("efter north ska status visa Tavern, fick %v", "after north the status should show Tavern, got %v", "north の後、status は Tavern であるべきですが、%v でした"), f)
	}
}

func TestJSONAndTextClientsTogether(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(mapTestWorld()))
	anna := jsonLogin(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	bertil.drain()

	bertil.send("say hi Anna")
	anna.expectText(fmt.Sprintf(MsgSays, "Bertil", "hi Anna"))
	anna.send("say hi Bertil")
	bertil.expectContains(fmt.Sprintf(MsgSays, "Anna", "hi Bertil"))

	here := strings.SplitN(MsgHere, "%s", 2)[0]
	bertil.send("look")
	for i := 0; i < 5; i++ {
		line, _ := bertil.readLine(testTimeout)
		if strings.HasPrefix(line, "{") {
			t.Fatalf(L("en telnet-klient ska aldrig få JSON, fick %q", "a telnet client must never get JSON, got %q", "telnet クライアントには JSON を送ってはいけませんが、%q を受信しました"), line)
		}
		if strings.HasPrefix(line, here) {
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
