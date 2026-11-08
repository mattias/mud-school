package mud

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

func TestAStarIsGeneric(t *testing.T) {
	// Rooms as strings: the short way round is a–b–e, not a–c–d–e.
	graph := map[string][]string{"a": {"b", "c"}, "b": {"e"}, "c": {"d"}, "d": {"e"}, "e": nil}
	path, ok := AStar("a", "e",
		func(n string) []string { return graph[n] },
		func(a, b string) int { return 1 },
		func(string) int { return 0 })
	if !ok || strings.Join(path, ",") != "b,e" {
		t.Fatalf(L("AStar på en graf av strängar: fick %v, %v; väntade [b e] (utan start, med mål)", "AStar on a graph of strings: got %v, %v; expected [b e] (without start, with goal)", "文字列のグラフでの AStar: %v, %v が返りました。[b e]（start を除き goal を含む）を期待しました"), path, ok)
	}
	// Costs matter: the direct edge is expensive.
	cost := map[[2]int]int{{0, 3}: 10, {0, 1}: 1, {1, 2}: 1, {2, 3}: 1}
	ints, ok := AStar(0, 3,
		func(n int) []int {
			var out []int
			for k := range cost {
				if k[0] == n {
					out = append(out, k[1])
				}
			}
			return out
		},
		func(a, b int) int { return cost[[2]int{a, b}] },
		func(int) int { return 0 })
	if !ok || fmt.Sprint(ints) != "[1 2 3]" {
		t.Errorf(L("AStar ska välja den billigaste vägen: fick %v", "AStar should choose the cheapest path: got %v", "AStar は最も安い経路を選ぶべきです: %v でした"), ints)
	}
	if _, ok := AStar(0, 9, func(int) []int { return nil }, func(a, b int) int { return 1 }, func(int) int { return 0 }); ok {
		t.Error(L("utan väg ska AStar returnera false", "without a path AStar should return false", "経路がなければ AStar は false を返すべきです"))
	}
	if p, ok := AStar(4, 4, func(int) []int { return nil }, func(a, b int) int { return 1 }, func(int) int { return 0 }); !ok || len(p) != 0 {
		t.Error(L("från ett mål till sig självt: en tom väg och true", "from a goal to itself: an empty path and true", "自分自身への経路: 空の経路と true"))
	}
}

func TestZonePath(t *testing.T) {
	z := zoneFrom(
		"###########",
		"#....#....#",
		"#....#....#",
		"#.........#",
		"###########",
	)
	from, to := Pos{1, 1}, Pos{9, 1}
	path, ok := z.Path(from, to, nil)
	if !ok || len(path) != 8 || path[len(path)-1] != to {
		t.Fatalf(L("Path runt väggen (diagonala steg tillåtna): %d steg %v; väntade 8 steg som slutar i %v", "Path round the wall (diagonal steps allowed): %d steps %v; expected 8 steps ending at %v", "壁を回り込む Path（斜め移動あり）: %d 歩 %v。%v で終わる 8 歩を期待しました"), len(path), path, to)
	}
	prev := from
	for _, p := range path {
		if chebyshev(prev, p) != 1 || !Walkable(z.At(p)) {
			t.Fatalf(L("varje steg ska vara till en gångbar granne: %v -> %v", "every step should be to a walkable neighbour: %v -> %v", "各歩は歩ける隣のマスへ: %v -> %v"), prev, p)
		}
		prev = p
	}
	blocked := func(p Pos) bool { return p == Pos{5, 3} }
	if _, ok := z.Path(from, to, blocked); ok {
		t.Error(L("blocked ska kunna stänga den enda vägen", "blocked should be able to close the only way", "blocked で唯一の道を閉じられるべきです"))
	}
}

func TestDijkstraMap(t *testing.T) {
	z := zoneFrom(
		"#######",
		"#.....#",
		"#.#####",
		"#.....#",
		"#######",
	)
	d := DijkstraMap(z, []Pos{{1, 1}})
	for p, want := range map[Pos]int{{1, 1}: 0, {5, 1}: 4, {1, 2}: 1, {2, 3}: 2, {5, 3}: 5} {
		if d[p] != want {
			t.Errorf(L("DijkstraMap[%v] = %d, väntade %d", "DijkstraMap[%v] = %d, expected %d", "DijkstraMap[%v] = %d。%d を期待しました"), p, d[p], want)
		}
	}
	if _, ok := d[Pos{3, 2}]; ok {
		t.Error(L("väggar ska inte finnas i kartan", "walls should not be in the map", "壁はマップに含めないべきです"))
	}
}

const huntWorldJSON = `{
  "start": "gate",
  "rooms": [
    {"id": "gate", "name": "Gate", "description": "A gate.", "terrain": "town", "exits": {"east": "field"},
     "zone": ["#####", "#...#", "#....", "#...#", "#####"]},
    {"id": "field", "name": "Field", "description": "Open.", "terrain": "town", "x": 1, "exits": {"west": "gate"},
     "zone": ["###############",
              "#.............#",
              "..............#",
              "#.............#",
              "###############"]}
  ],
  "npcs": [
    {"id": "wolf", "name": "a wolf", "description": "Hungry.", "keywords": ["wolf"], "room": "field",
     "hp": 1000, "damage": [1, 1], "attackable": true, "behaviors": [{"type": "hunt", "range": 8}]}
  ]
}`

type huntTest struct {
	srv  *Server
	addr string
}

func newHuntTest(t *testing.T, world string) huntTest {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(world))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, Rand: rand.New(rand.NewPCG(1, 2))})
	return huntTest{srv, startServer(t, srv)}
}

func (h huntTest) tick() {
	h.srv.call(func() {
		h.srv.ticks++
		h.srv.onTick()
	})
}

func (h huntTest) npcPos(id string) Pos {
	var p Pos
	h.srv.call(func() { p = h.srv.world.NPCs[id].Pos })
	return p
}

func TestHunterComes(t *testing.T) {
	h := newHuntTest(t, huntWorldJSON)
	c := login(t, h.addr, "Anna")
	c.drain()
	c.send("go east") // Anna arrives at (1,2); the wolf waits at (7,2)
	c.drain()
	start := h.npcPos("wolf")
	if start != (Pos{7, 2}) {
		t.Fatalf("the wolf should start in the middle (7,2), got %v", start)
	}
	h.tick()
	if p := h.npcPos("wolf"); chebyshev(p, Pos{1, 2}) != 5 {
		t.Fatalf(L("på ett tick ska vargen ta ett steg mot Anna: %v -> %v", "in one tick the wolf should take one step towards Anna: %v -> %v", "1 ティックで狼はアンナに向かって 1 歩進むべきです: %v -> %v"), start, p)
	}
	for i := 0; i < 4; i++ {
		h.tick()
	}
	if p := h.npcPos("wolf"); chebyshev(p, Pos{1, 2}) != 1 {
		t.Fatalf(L("efter fem tick ska vargen stå bredvid Anna, står på %v", "after five ticks the wolf should stand next to Anna, it is at %v", "5 ティック後、狼はアンナの隣にいるべきですが %v にいます"), p)
	}
	c.drain()
	h.tick()
	c.expectContains(fmt.Sprintf(MsgAttacksYou, "a wolf"))
	var fighting bool
	h.srv.call(func() { fighting = h.srv.players["Anna"].target != nil })
	if !fighting {
		t.Error(L("när vargen anfaller ska Anna vara i strid med den (p.target)", "when the wolf attacks, Anna should be fighting it (p.target)", "狼が襲ってきたら、アンナはそれと戦闘中（p.target）であるべきです"))
	}
}

const walledWorldJSON = `{
  "start": "gate",
  "rooms": [
    {"id": "gate", "name": "Gate", "description": "A gate.", "terrain": "town", "exits": {"east": "field"},
     "zone": ["#####", "#...#", "#....", "#...#", "#####"]},
    {"id": "field", "name": "Field", "description": "Walled.", "terrain": "town", "x": 1, "exits": {"west": "gate"},
     "zone": ["###############",
              "#...#.........#",
              "....#.........#",
              "#...#.........#",
              "###############"]}
  ],
  "npcs": [
    {"id": "wolf", "name": "a wolf", "description": "Hungry.", "keywords": ["wolf"], "room": "field",
     "hp": 1000, "damage": [1, 1], "attackable": true, "behaviors": [{"type": "hunt", "range": 8}]}
  ]
}`

func TestHunterNeedsToSee(t *testing.T) {
	h := newHuntTest(t, walledWorldJSON)
	c := login(t, h.addr, "Anna")
	c.drain()
	c.send("go east")
	c.drain()
	start := h.npcPos("wolf")
	for i := 0; i < 3; i++ {
		h.tick()
	}
	if p := h.npcPos("wolf"); p != start {
		t.Errorf(L("vargen ser inte Anna genom väggen och ska stå still, men gick %v -> %v", "the wolf can't see Anna through the wall and should stay put, but walked %v -> %v", "狼は壁越しにアンナを見られないので動かないはずですが、%v -> %v と歩きました"), start, p)
	}
}

func TestFleerRunsAway(t *testing.T) {
	world := strings.Replace(huntWorldJSON, `"behaviors": [{"type": "hunt", "range": 8}]`, `"behaviors": [{"type": "flee", "below": 0.5}]`, 1)
	h := newHuntTest(t, world)
	c := login(t, h.addr, "Anna")
	c.drain()
	c.send("go east")
	c.drain()
	for i := 0; i < 5; i++ {
		c.send("e") // Anna walks up to the wolf at (7,2)
	}
	c.drain()
	h.tick()
	if p := h.npcPos("wolf"); p != (Pos{7, 2}) {
		t.Fatalf(L("med full hälsa ska vargen inte fly, men gick till %v", "with full health the wolf should not flee, but went to %v", "体力が満タンなら狼は逃げないはずですが、%v へ行きました"), p)
	}
	h.srv.call(func() { h.srv.world.NPCs["wolf"].hp = 10 })
	before := chebyshev(h.npcPos("wolf"), Pos{6, 2})
	h.tick()
	h.tick()
	if after := chebyshev(h.npcPos("wolf"), Pos{6, 2}); after <= before {
		t.Errorf(L("med låg hälsa ska vargen springa ifrån Anna (Dijkstra-karta): avstånd %d -> %d", "with low health the wolf should run away from Anna (Dijkstra map): distance %d -> %d", "体力が少なければ狼はアンナから逃げるべきです（ダイクストラマップ）: 距離 %d -> %d"), before, after)
	}
}
