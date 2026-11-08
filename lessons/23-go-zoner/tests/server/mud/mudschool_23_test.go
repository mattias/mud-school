package mud

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// A small world: the square has a hand-made zone, the hall and the loft get
// plain generated zones of 9×5.
const zoneWorldJSON = `{
  "start": "torg",
  "zone_size": [9, 5],
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town",
     "exits": {"north": "hall"},
     "zone": ["####.####",
              "#.......#",
              "#..~....#",
              "#.......#",
              "#########"]},
    {"id": "hall", "name": "Hall", "description": "A hall.", "terrain": "inside", "y": -1,
     "exits": {"south": "torg", "up": "loft"}},
    {"id": "loft", "name": "Loft", "description": "A loft.", "terrain": "inside", "y": -1, "z": 1,
     "exits": {"down": "hall"}}
  ],
  "items": [{"id": "lamp", "name": "a lamp", "description": "Brass.", "keywords": ["lamp"], "room": "torg"}],
  "npcs": [{"id": "cat", "name": "a cat", "description": "Black.", "keywords": ["cat"], "room": "hall"}]
}`

func newZoneServer(t *testing.T) string {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(zoneWorldJSON))
	if err != nil {
		t.Fatalf(L("världen med zoner gick inte att läsa: %v", "the world with zones could not be loaded: %v", "ゾーン付きの世界を読み込めませんでした: %v"), err)
	}
	return startServer(t, NewServerWithOptions(Options{World: w, Tick: time.Hour}))
}

func TestZoneFrameOnLogin(t *testing.T) {
	c := loginJSON(t, newZoneServer(t), "Anna")
	z := c.expectZone("torg")
	if z["width"] != 9.0 || z["height"] != 5.0 {
		t.Fatalf(L("zonens storlek: %v×%v, väntade 9×5", "the zone's size: %v×%v, expected 9×5", "ゾーンの大きさ: %v×%v。9×5 を期待しました"), z["width"], z["height"])
	}
	rows := fmt.Sprint(z["rows"])
	if rows != "[####.#### #.......# #..~....# #.......# #########]" {
		t.Errorf(L("rows ska vara den handgjorda zonen, fick %s", "rows should be the hand-made zone, got %s", "rows は手作りのゾーンであるべきですが %s でした"), rows)
	}
	legend, _ := z["legend"].(map[string]any)
	if legend["#"] != "wall" || legend["."] != "floor" || legend["~"] != "water" {
		t.Errorf(L("legend saknar betydelser: %v", "the legend lacks meanings: %v", "legend に意味が足りません: %v"), legend)
	}
	if z["self"] != "p:Anna" {
		t.Errorf(`self = %v, expected "p:Anna"`, z["self"])
	}
	ents := entitiesOf(t, z["entities"])
	if e := ents["p:Anna"]; e.kind != "player" || e.glyph != "@" || e.x != 4 || e.y != 2 {
		t.Errorf(L("Anna ska stå mitt i zonen (4,2) som en player med glyph @, fick %+v", "Anna should stand in the middle of the zone (4,2) as a player with glyph @, got %+v", "アンナはゾーンの中央 (4,2) に glyph @ の player として立つべきですが %+v でした"), e)
	}
	if e := ents["i:lamp"]; e.kind != "item" || e.glyph != "*" {
		t.Errorf(L("lampan ska vara en entity i:lamp av sorten item med glyph *, fick %+v", "the lamp should be an entity i:lamp of kind item with glyph *, got %+v", "ランプは種類 item、glyph * のエンティティ i:lamp であるべきですが %+v でした"), e)
	}
}

func TestStepAndBlockedInTelnet(t *testing.T) {
	c := login(t, newZoneServer(t), "Anna")
	c.drain()
	c.send("e")
	var rows []string
	for i := 0; i < 5; i++ {
		line, _ := c.readLine(testTimeout)
		rows = append(rows, line)
	}
	want := []string{"####.####", "#.......#", "#..~*@..#", "#.......#", "#########"}
	if strings.Join(rows, "\n") != strings.Join(want, "\n") {
		t.Fatalf(L("efter ett steg österut ska telnet få zonen ritad med @ och lampan (*):\n  fick:\n%s\n  väntade:\n%s",
			"after a step east, telnet should get the zone drawn with @ and the lamp (*):\n  got:\n%s\n  expected:\n%s",
			"東へ 1 歩進んだら、telnet には @ とランプ（*）を描いたゾーンが届くべきです:\n  実際:\n%s\n  期待:\n%s"), strings.Join(rows, "\n"), strings.Join(want, "\n"))
	}
	c.send("w")
	c.drain()
	c.send("w") // water
	c.expect(MsgBlocked)
	c.send("s")
	c.drain()
	c.send("s") // wall
	c.expect(MsgBlocked)
	c.send("se") // diagonal into the wall too
	c.expect(MsgBlocked)
}

func TestWalkOutThroughTheOpening(t *testing.T) {
	c := loginJSON(t, newZoneServer(t), "Anna")
	c.expectZone("torg")
	for i := 0; i < 2; i++ {
		c.send("n")
		if d := c.expectDelta(); entitiesOf(t, d["moved"])["p:Anna"].y != 1-i {
			t.Fatalf(L("ett steg norrut ska ge en delta där Anna har flyttat, fick %v", "a step north should give a delta where Anna has moved, got %v", "北へ 1 歩進むと、アンナが移動した delta が届くべきですが %v でした"), d)
		}
	}
	c.send("n") // out through the opening at the top
	c.expectContains("Hall")
	z := c.expectZone("hall")
	if e := entitiesOf(t, z["entities"])["p:Anna"]; e.x != 4 || e.y != 3 {
		t.Errorf(L("i hallen ska Anna komma in ett steg innanför den södra öppningen (4,3), fick %+v",
			"in the hall Anna should arrive one step inside the southern opening (4,3), got %+v",
			"ホールではアンナは南の開口部の 1 歩内側 (4,3) に現れるべきですが %+v でした"), e)
	}
	if e := entitiesOf(t, z["entities"])["n:cat"]; e.kind != "creature" || e.glyph != "c" {
		t.Errorf(L("katten ska vara en creature med glyph c (första bokstaven i id), fick %+v", "the cat should be a creature with glyph c (the first letter of its id), got %+v", "猫は glyph c（id の最初の文字）の creature であるべきですが %+v でした"), e)
	}
	if rows := fmt.Sprint(z["rows"]); !strings.HasPrefix(rows, "[######### #...<...# ") || !strings.HasSuffix(rows, " ####.####]") {
		t.Errorf(L("hallens genererade zon ska ha väggar, en trappa upp (<) och en öppning i söder: %s",
			"the hall's generated zone should have walls, stairs up (<) and an opening to the south: %s",
			"ホールの生成されたゾーンには壁、上り階段（<）、南の開口部があるべきです: %s"), rows)
	}
}

func TestGoAndStairs(t *testing.T) {
	c := login(t, newZoneServer(t), "Anna")
	c.drain()
	c.send("go")
	c.expect(MsgGoWhere)
	c.send("go west")
	c.expect(MsgCantGo)
	c.send("up")
	c.expect(MsgNoStairs)
	c.send("go n")
	c.expect("Hall")
	c.drain()
	c.send("n") // (4,3) -> (4,2)
	c.drain()
	c.send("n") // the stairs at (4,1)
	c.drain()
	c.send("up")
	c.expect("Loft")
	c.drain()
	c.send("zone")
	var rows []string
	for i := 0; i < 5; i++ {
		line, _ := c.readLine(testTimeout)
		rows = append(rows, line)
	}
	if rows[3] != "#...@...#" {
		t.Errorf(L("uppe på loftet ska Anna stå på trappan ner (4,3):\n%s", "up in the loft Anna should stand on the stairs down (4,3):\n%s", "屋根裏ではアンナは下り階段 (4,3) の上に立つべきです:\n%s"), strings.Join(rows, "\n"))
	}
}

func TestOthersSeeDeltas(t *testing.T) {
	addr := newZoneServer(t)
	anna := loginJSON(t, addr, "Anna")
	anna.expectZone("torg")
	anna.drain()
	bertil := loginJSON(t, addr, "Bertil")
	bertil.expectZone("torg")
	if d := anna.expectDelta(); entitiesOf(t, d["added"])["p:Bertil"].kind != "player" {
		t.Fatalf(L("när Bertil kommer in ska Anna få en delta med added p:Bertil, fick %v", "when Bertil arrives Anna should get a delta with added p:Bertil, got %v", "ベルティルが来たら、アンナには added に p:Bertil を含む delta が届くべきですが %v でした"), d)
	}
	anna.drain()
	bertil.send("e")
	if d := anna.expectDelta(); entitiesOf(t, d["moved"])["p:Bertil"].x != 5 {
		t.Fatalf(L("när Bertil går österut ska Anna få en delta med moved p:Bertil x=5, fick %v", "when Bertil walks east Anna should get a delta with moved p:Bertil x=5, got %v", "ベルティルが東へ歩いたら、アンナには moved に p:Bertil x=5 を含む delta が届くべきですが %v でした"), d)
	}
	anna.drain()
	bertil.send("go north")
	if d := anna.expectDelta(); fmt.Sprint(d["removed"]) != "[p:Bertil]" {
		t.Fatalf(L("när Bertil går därifrån ska Anna få en delta med removed [p:Bertil], fick %v", "when Bertil leaves Anna should get a delta with removed [p:Bertil], got %v", "ベルティルが去ったら、アンナには removed [p:Bertil] の delta が届くべきですが %v でした"), d)
	}
}

func TestZoneValidation(t *testing.T) {
	bad := strings.Replace(zoneWorldJSON, `"####.####",`, `"####.###",`, 1)
	if _, err := LoadWorld(strings.NewReader(bad)); err == nil || !strings.Contains(err.Error(), strings.SplitN(MsgErrZoneRows, "%", 2)[0]) {
		t.Errorf(L("rader med olika längd ska ge MsgErrZoneRows, fick %v", "rows of different lengths should give MsgErrZoneRows, got %v", "長さの違う行は MsgErrZoneRows になるべきですが %v でした"), err)
	}
	closed := strings.Replace(zoneWorldJSON, `"####.####",`, `"#########",`, 1)
	if _, err := LoadWorld(strings.NewReader(closed)); err == nil || !strings.Contains(err.Error(), fmt.Sprintf(MsgErrZoneExit, "torg", "north")) {
		t.Errorf(L("en utgång norrut utan öppning i zonen ska ge MsgErrZoneExit, fick %v", "an exit north without an opening in the zone should give MsgErrZoneExit, got %v", "ゾーンに開口部のない北の出口は MsgErrZoneExit になるべきですが %v でした"), err)
	}
}
