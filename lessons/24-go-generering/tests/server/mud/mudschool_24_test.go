package mud

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
)

func grid(rows ...string) [][]bool {
	g := make([][]bool, len(rows))
	for y, row := range rows {
		for _, c := range row {
			g[y] = append(g[y], c == '#')
		}
	}
	return g
}

func gridString(g [][]bool) string {
	var b strings.Builder
	for _, row := range g {
		for _, wall := range row {
			if wall {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestCAStep(t *testing.T) {
	in := grid(
		"#####",
		"#.#.#",
		"#...#",
		"#.#.#",
		"#####",
	)
	want := "#####\n##.##\n##.##\n##.##\n#####\n"
	if got := gridString(caStep(in)); got != want {
		t.Fatalf(L("caStep (4-5-regeln, rutor utanför räknas som vägg):\n  fick:\n%s  väntade:\n%s",
			"caStep (the 4-5 rule, cells outside count as walls):\n  got:\n%s  expected:\n%s",
			"caStep（4-5 ルール、外側のマスは壁として数える）:\n  実際:\n%s  期待:\n%s"), got, want)
	}
}

func TestValueNoise(t *testing.T) {
	for i := 0; i < 1000; i++ {
		x, y := float64(i)*0.37, float64(i%17)*0.91
		n := ValueNoise(7, x, y)
		if n < 0 || n >= 1 {
			t.Fatalf("ValueNoise(7, %v, %v) = %v – %s", x, y, n, L("utanför [0,1)", "outside [0,1)", "[0,1) の範囲外"))
		}
		if n != ValueNoise(7, x, y) {
			t.Fatal(L("ValueNoise ska ge samma värde för samma frö och punkt", "ValueNoise should give the same value for the same seed and point", "ValueNoise は同じシードと点に対して同じ値を返すべきです"))
		}
		if d := n - ValueNoise(7, x+0.01, y); d > 0.05 || d < -0.05 {
			t.Fatalf(L("ValueNoise ska vara mjukt: (%v,%v) och 0,01 bort skiljer %v", "ValueNoise should be smooth: (%v,%v) and 0.01 away differ by %v", "ValueNoise はなめらかであるべきです: (%v,%v) と 0.01 離れた点の差が %v"), x, y, d)
		}
	}
	same := 0
	for i := 0; i < 100; i++ {
		if ValueNoise(1, float64(i)+0.5, 0.5) == ValueNoise(2, float64(i)+0.5, 0.5) {
			same++
		}
	}
	if same > 5 {
		t.Error(L("olika frön ska ge olika brus", "different seeds should give different noise", "シードが違えばノイズも違うべきです"))
	}
}

func TestDrunkardWalk(t *testing.T) {
	z := NewZone(21, 11, TileWall)
	from, to := Pos{2, 2}, Pos{18, 8}
	drunkardWalk(z, from, to, rand.New(rand.NewPCG(1, 2)), TileFloor)
	if !reachable(z, from)[to] {
		t.Fatalf(L("drunkardWalk ska gräva en gång hela vägen fram:\n%s", "drunkardWalk should carve a tunnel all the way:\n%s", "drunkardWalk は最後までトンネルを掘るべきです:\n%s"), strings.Join(z.Rows(), "\n"))
	}
	for x := 0; x < z.W; x++ {
		if z.At(Pos{x, 0}) != TileWall || z.At(Pos{x, z.H - 1}) != TileWall {
			t.Fatalf(L("drunkardWalk får inte gräva i zonens kant:\n%s", "drunkardWalk must not dig into the edge of the zone:\n%s", "drunkardWalk はゾーンの端を掘ってはいけません:\n%s"), strings.Join(z.Rows(), "\n"))
		}
	}
}

func allRoom(terrain string) *Room {
	return &Room{ID: "test-" + terrain, Terrain: terrain, Exits: map[string]string{
		"north": "a", "south": "b", "east": "c", "west": "d", "up": "e", "down": "f",
	}}
}

// checkConnected: every walkable cell can be reached from every opening.
func checkConnected(t *testing.T, r *Room, z *Zone) {
	t.Helper()
	show := strings.Join(z.Rows(), "\n")
	var start Pos
	for dir := range r.Exits {
		p, ok := z.opening(dir)
		if !ok {
			t.Fatalf(L("%s: ingen öppning för %s:\n%s", "%s: no opening for %s:\n%s", "%s: %s の開口部がありません:\n%s"), r.ID, dir, show)
		}
		start = p
	}
	walkable := 0
	for y := 0; y < z.H; y++ {
		for x := 0; x < z.W; x++ {
			if Walkable(z.At(Pos{x, y})) {
				walkable++
			}
		}
	}
	if n := len(reachable(z, start)); n != walkable {
		t.Fatalf(L("%s: bara %d av %d gångbara rutor går att nå – zonen måste hänga ihop:\n%s",
			"%s: only %d of %d walkable cells can be reached – the zone must be connected:\n%s",
			"%s: 歩ける %[3]d マスのうち到達できるのは %[2]d マスだけです。ゾーンはつながっている必要があります:\n%[4]s"), r.ID, n, walkable, show)
	}
	for x := 0; x < z.W; x++ {
		for _, y := range []int{0, z.H - 1} {
			if Walkable(z.At(Pos{x, y})) && x != z.W/2 {
				t.Fatalf(L("%s: kanten får bara vara öppen vid utgångarna:\n%s", "%s: the edge may only be open at the exits:\n%s", "%s: 端が開いてよいのは出口だけです:\n%s"), r.ID, show)
			}
		}
	}
}

func TestGeneratedZones(t *testing.T) {
	for _, terrain := range Terrains {
		r := allRoom(terrain)
		for seed := uint64(1); seed <= 20; seed++ {
			z := GenerateZone(r, 31, 13, seed)
			if z.W != 31 || z.H != 13 {
				t.Fatalf("GenerateZone: %d×%d", z.W, z.H)
			}
			checkConnected(t, r, z)
		}
	}
}

func count(z *Zone, t byte) int {
	n := 0
	for _, row := range z.Tiles {
		for _, c := range row {
			if c == t {
				n++
			}
		}
	}
	return n
}

func TestTerrainsLookDifferent(t *testing.T) {
	gen := func(terrain string) *Zone { return GenerateZone(allRoom(terrain), 41, 17, 99) }
	cave := gen("cave")
	inner := 39 * 15
	if walls := count(cave, TileWall) - 2*41 - 2*15; walls < inner/5 || walls > inner*3/4 {
		t.Errorf(L("en grotta ska vara en blandning av vägg och golv (cellautomat), fick %d väggar av %d:\n%s",
			"a cave should be a mix of wall and floor (cellular automaton), got %d walls of %d:\n%s",
			"洞窟は壁と床の混ざったもの（セル・オートマトン）であるべきですが、壁 %d / %d でした:\n%s"), walls, inner, strings.Join(cave.Rows(), "\n"))
	}
	checks := []struct {
		terrain string
		tile    byte
	}{{"forest", TileTree}, {"forest", TileGrass}, {"field", TileGrass}, {"mountain", TileRock}, {"water", TileWater}, {"town", TileFloor}}
	for _, c := range checks {
		if count(gen(c.terrain), c.tile) == 0 {
			t.Errorf(L("terrängen %s ska innehålla rutan %q", "the terrain %s should contain the tile %q", "地形 %s にはタイル %q が含まれるべきです"), c.terrain, c.tile)
		}
	}
	if count(gen("town"), TileTree) != 0 {
		t.Error(L("en stad ska inte ha träd", "a town should have no trees", "町に木があってはいけません"))
	}
}

func TestSameSeedSameZone(t *testing.T) {
	r := allRoom("cave")
	a := strings.Join(GenerateZone(r, 31, 13, 5).Rows(), "\n")
	b := strings.Join(GenerateZone(r, 31, 13, 5).Rows(), "\n")
	c := strings.Join(GenerateZone(r, 31, 13, 6).Rows(), "\n")
	other := strings.Join(GenerateZone(&Room{ID: "other", Terrain: "cave", Exits: r.Exits}, 31, 13, 5).Rows(), "\n")
	if a != b {
		t.Fatal(L("samma rum och frö ska ge exakt samma zon", "the same room and seed should give exactly the same zone", "同じ部屋とシードはまったく同じゾーンになるべきです"))
	}
	if a == c || a == other {
		t.Error(L("ett annat frö eller ett annat rum ska ge en annan zon", "another seed or another room should give another zone", "シードや部屋が違えば別のゾーンになるべきです"))
	}
}

func TestWorldSeed(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(`{"start": "g", "seed": 42, "zone_size": [25, 11],
	  "rooms": [{"id": "g", "name": "Grotto", "description": "Dark.", "terrain": "cave"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(w.zoneOf(w.Room("g")).Rows(), "\n")
	want := strings.Join(GenerateZone(w.Room("g"), 25, 11, 42).Rows(), "\n")
	if got != want {
		t.Errorf(L("rum utan handgjord zon ska få GenerateZone med världens \"seed\" och \"zone_size\":\n%s",
			"rooms without a hand-made zone should get GenerateZone with the world's \"seed\" and \"zone_size\":\n%s",
			"手作りのゾーンがない部屋には、世界の \"seed\" と \"zone_size\" で GenerateZone したゾーンを使うべきです:\n%s"), got)
	}
}

func TestDefaultWorldIsWalkable(t *testing.T) {
	w := DefaultWorld()
	for _, r := range w.Rooms {
		z := w.zoneOf(r)
		for dir := range r.Exits {
			if _, ok := z.opening(dir); !ok {
				t.Errorf(fmt.Sprintf(MsgErrZoneExit, r.ID, dir))
			}
		}
		if len(r.Exits) > 0 {
			checkConnectedSimple(t, r, z)
		}
	}
}

func checkConnectedSimple(t *testing.T, r *Room, z *Zone) {
	t.Helper()
	var start Pos
	for dir := range r.Exits {
		start, _ = z.opening(dir)
	}
	reach := reachable(z, start)
	for dir := range r.Exits {
		if p, _ := z.opening(dir); !reach[p] {
			t.Errorf(L("%s: öppningen %s går inte att nå:\n%s", "%s: the opening %s can't be reached:\n%s", "%s: 開口部 %s に到達できません:\n%s"), r.ID, dir, strings.Join(z.Rows(), "\n"))
		}
	}
}
