package mud

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// showFOV draws a field of view: o = the origin, * = visible, . = not.
func showFOV(z *Zone, origin Pos, vis map[Pos]bool) string {
	var b strings.Builder
	for y := 0; y < z.H; y++ {
		for x := 0; x < z.W; x++ {
			p := Pos{x, y}
			switch {
			case p == origin:
				b.WriteByte('o')
			case vis[p]:
				b.WriteByte('*')
			default:
				b.WriteByte(z.At(p))
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestFOVPillar(t *testing.T) {
	z := zoneFrom(
		"#####################",
		"#...................#",
		"#...................#",
		"#...................#",
		"#...................#",
		"#...........#.......#",
		"#...................#",
		"#...................#",
		"#...................#",
		"#...................#",
		"#####################",
	)
	origin := Pos{10, 5}
	vis := ComputeFOV(z, origin, PlayerSight)
	show := showFOV(z, origin, vis)
	for _, c := range []struct {
		p    Pos
		want bool
	}{
		{origin, true}, {Pos{11, 5}, true}, {Pos{12, 5}, true}, // the pillar itself is seen
		{Pos{13, 5}, false}, {Pos{15, 5}, false}, // the shadow behind it
		{Pos{10, 1}, true}, {Pos{4, 8}, true}, {Pos{14, 2}, true},
		{Pos{10, 0}, true},  // the wall at the edge is seen
		{Pos{19, 5}, false}, // behind the pillar and too far (9 > 8)
		{Pos{1, 1}, false},  // too far: 9² + 4² > 8²
	} {
		if vis[c.p] != c.want {
			t.Fatalf(L("ComputeFOV: %v ska vara synlig=%v (o = du, * = synligt):\n%s", "ComputeFOV: %v should be visible=%v (o = you, * = visible):\n%s", "ComputeFOV: %v は visible=%v であるべきです（o = 自分、* = 見える）:\n%s"), c.p, c.want, show)
		}
	}
}

func TestFOVWalls(t *testing.T) {
	z := zoneFrom(
		"###########",
		"#....#....#",
		"#....#....#",
		"#....#....#",
		"###########",
	)
	origin := Pos{2, 2}
	vis := ComputeFOV(z, origin, PlayerSight)
	show := showFOV(z, origin, vis)
	if !vis[Pos{5, 2}] || !vis[Pos{0, 0}] {
		t.Fatalf(L("väggarna runt dig ska synas:\n%s", "the walls around you should be visible:\n%s", "周りの壁は見えるべきです:\n%s"), show)
	}
	for x := 6; x < 10; x++ {
		for y := 1; y < 4; y++ {
			if vis[Pos{x, y}] {
				t.Fatalf(L("inget bakom väggen får synas:\n%s", "nothing behind the wall may be visible:\n%s", "壁の向こうは何も見えてはいけません:\n%s"), show)
			}
		}
	}
	// A tree blocks, water doesn't.
	z = zoneFrom("#######", "#..T..#", "#######")
	if ComputeFOV(z, Pos{1, 1}, 8)[Pos{5, 1}] {
		t.Error(L("ett träd ska skymma sikten", "a tree should block the view", "木は視界をさえぎるべきです"))
	}
	z = zoneFrom("#######", "#..~..#", "#######")
	if !ComputeFOV(z, Pos{1, 1}, 8)[Pos{5, 1}] {
		t.Error(L("man ska se över vatten", "you should see across water", "水の向こうは見えるべきです"))
	}
}

const fovWorldJSON = `{
  "start": "hus",
  "rooms": [{"id": "hus", "name": "House", "description": "Two rooms.", "terrain": "inside",
    "zone": ["###############",
             "#.....#.......#",
             "#.....#.......#",
             "#.....#.......#",
             "#.............#",
             "#.....#.......#",
             "###############"]}]
}`

func newFOVServer(t *testing.T) string {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(fovWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	return startServer(t, NewServerWithOptions(Options{World: w, Tick: time.Hour}))
}

func TestZoneShowsOnlyExplored(t *testing.T) {
	c := loginJSON(t, newFOVServer(t), "Anna") // Anna stands at (7,3)
	z := c.expectZone("hus")
	rows, _ := z["rows"].([]any)
	vis, _ := z["visible"].([]any)
	if len(rows) != 7 || len(vis) != 7 {
		t.Fatalf(L("zone-framen ska ha rows och visible, 7 rader var", "the zone frame should have rows and visible, 7 rows each", "zone フレームには 7 行ずつの rows と visible があるべきです"))
	}
	row1, vis1 := fmt.Sprint(rows[1]), fmt.Sprint(vis[1])
	if row1[1] != ' ' || vis1[1] != '0' {
		t.Errorf(L("det vänstra rummet bakom väggen ska vara okänt (' ') och osynligt ('0'): rows[1]=%q visible[1]=%q",
			"the left room behind the wall should be unknown (' ') and not visible ('0'): rows[1]=%q visible[1]=%q",
			"壁の向こうの左の部屋は未知（' '）で見えない（'0'）べきです: rows[1]=%q visible[1]=%q"), row1, vis1)
	}
	if row1[8] != '.' || vis1[8] != '1' {
		t.Errorf(L("rummet du står i ska synas: rows[1]=%q visible[1]=%q", "the room you stand in should be visible: rows[1]=%q visible[1]=%q", "自分のいる部屋は見えるべきです: rows[1]=%q visible[1]=%q"), row1, vis1)
	}

	c.send("s") // (7,4): now Anna looks through the doorway
	d := c.expectDelta()
	tiles, _ := d["tiles"].([]any)
	revealed := false
	for _, raw := range tiles {
		tc, _ := raw.(map[string]any)
		if x, _ := tc["x"].(float64); x < 6 && tc["code"] == "." && tc["meaning"] == "floor" {
			revealed = true
		}
	}
	if !revealed {
		t.Fatalf(L("när Anna ser genom dörröppningen ska deltan ha tiles för det vänstra rummet ({x, y, code, meaning}), fick %v",
			"when Anna looks through the doorway the delta should have tiles for the left room ({x, y, code, meaning}), got %v",
			"アンナが戸口から覗いたら、delta に左の部屋の tiles（{x, y, code, meaning}）が含まれるべきですが %v でした"), d)
	}
	if _, ok := d["visible"]; !ok {
		t.Errorf(L("synfältet har ändrats – deltan ska ha visible", "the field of view has changed – the delta should have visible", "視界が変わったので delta に visible があるべきです"))
	}
}

func TestOthersHiddenBehindWalls(t *testing.T) {
	addr := newFOVServer(t)
	anna := loginJSON(t, addr, "Anna")
	anna.expectZone("hus")
	bertil := loginJSON(t, addr, "Bertil")
	bertil.expectZone("hus")
	anna.drain()
	// Bertil walks into the left room: (7,3) -> (7,4) -> (6,4) -> (5,4) -> (4,3) -> (3,2) -> (2,1)
	for _, step := range []string{"s", "w", "w", "nw", "nw", "nw"} {
		bertil.send(step)
		bertil.drain()
	}
	removed := false
	for i := 0; i < 8 && !removed; i++ {
		d := anna.expectDelta()
		removed = strings.Contains(fmt.Sprint(d["removed"]), "p:Bertil")
	}
	if !removed {
		t.Fatalf(L("när Bertil går in bakom väggen ska Anna få removed p:Bertil", "when Bertil walks in behind the wall Anna should get removed p:Bertil", "ベルティルが壁の向こうへ入ったら、アンナには removed p:Bertil が届くべきです"))
	}
	anna.drain()
	anna.send("zone")
	z := anna.expectZone("hus")
	if _, seen := entitiesOf(t, z["entities"])["p:Bertil"]; seen {
		t.Errorf(L("Bertil syns inte – han ska inte heller finnas i zone-framen", "Bertil can't be seen – he must not be in the zone frame either", "ベルティルは見えないので、zone フレームにも含めてはいけません"))
	}
	if _, me := entitiesOf(t, z["entities"])["p:Anna"]; !me {
		t.Errorf(L("man ser alltid sig själv", "you always see yourself", "自分自身は常に見えます"))
	}
}

func TestTelnetSeesOnlyExplored(t *testing.T) {
	c := login(t, newFOVServer(t), "Anna")
	c.drain()
	c.send("zone")
	line0, _ := c.readLine(testTimeout)
	line1, _ := c.readLine(testTimeout)
	if line1 != "      #.......#" {
		t.Errorf(L("telnet ska bara se det utforskade:\n%s\n%s", "telnet should only see what has been explored:\n%s\n%s", "telnet には探索済みの部分だけが見えるべきです:\n%s\n%s"), line0, line1)
	}
}
