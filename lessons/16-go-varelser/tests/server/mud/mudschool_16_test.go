package mud

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const npcWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "exits": {"north": "krog"}},
    {"id": "krog", "name": "Tavern", "description": "A tavern.", "terrain": "inside", "y": -1, "exits": {"south": "torg"}},
    {"id": "cell", "name": "Cell", "description": "No way out.", "terrain": "inside", "x": 5}
  ],
  "npcs": [
    {"id": "cat", "name": "a black cat", "description": "It stares at you.", "keywords": ["cat"], "room": "torg",
     "behaviors": [{"type": "wander", "chance": 1}]},
    {"id": "parrot", "name": "a parrot", "description": "Green and loud.", "keywords": ["parrot", "bird"], "room": "cell",
     "behaviors": [{"type": "talk", "chance": 1, "lines": ["Pretty bird!"]}]}
  ]
}`

func npcWorld(t *testing.T) *World {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(npcWorldJSON))
	if err != nil {
		t.Fatalf(L("världen med varelser gick inte att läsa: %v", "the world with creatures could not be loaded: %v", "生き物付きの世界を読み込めませんでした: %v"), err)
	}
	return w
}

func seeded() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func TestLoadNPCsErrors(t *testing.T) {
	tests := []struct{ name, npcs, want string }{
		{"missing room", `{"npcs": [{"id": "x", "name": "X", "room": "nowhere"}]}`, `"nowhere"`},
		{"twice", `{"npcs": [{"id": "x", "name": "X", "room": "torg"}, {"id": "x", "name": "Y", "room": "torg"}]}`, `"x"`},
		{"no name", `{"npcs": [{"id": "x", "room": "torg"}]}`, `"x"`},
		{"unknown behaviour", `{"npcs": [{"id": "x", "name": "X", "room": "torg", "behaviors": [{"type": "dance"}]}]}`, `"dance"`},
		{"typo", `{"npcs": [{"id": "x", "nmae": "X", "room": "torg"}]}`, ""},
	}
	for _, tt := range tests {
		w := npcWorld(t)
		err := w.LoadNPCs(strings.NewReader(tt.npcs))
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf(L("%s: LoadNPCs ska ge ett fel som nämner %s, gav %v", "%s: LoadNPCs should return an error mentioning %s, returned %v", "%s: LoadNPCs は %s を含むエラーを返すべきですが、%v でした"), tt.name, tt.want, err)
		}
	}
}

func TestDefaultWorldHasNPCs(t *testing.T) {
	w := DefaultWorld()
	if len(w.NPCs) < 3 {
		t.Errorf(L("standardvärlden har %d varelser – minst 3 i npcs.json, tack!", "the default world has %d creatures – at least 3 in npcs.json, please!", "標準の世界の生き物は %d 匹です。npcs.json には 3 匹以上入れてください！"), len(w.NPCs))
	}
	for id, n := range w.NPCs {
		if n.Location == nil || n.Location.ID != n.Room {
			t.Errorf(L("varelsen %q ska stå i sitt startrum %q", "the creature %q should be in its start room %q", "生き物 %q は開始部屋 %q にいるべきです"), id, n.Room)
		}
	}
}

func TestNPCInLookAndExamine(t *testing.T) {
	// A frozen world (tick of one hour): the cat stays where it is.
	srv := NewServerWithOptions(Options{World: npcWorld(t), Tick: time.Hour, Rand: seeded()})
	addr := startServer(t, srv)
	c := login(t, addr, "Anna")
	c.expectContains(fmt.Sprintf(MsgHere, "a black cat"))
	c.send("examine cat")
	c.expect("It stares at you.")
	c.send("get cat")
	c.expect(fmt.Sprintf(MsgNoSuchItem, "cat"))
}

func TestPlayersAndNPCsInLook(t *testing.T) {
	srv := NewServerWithOptions(Options{World: npcWorld(t), Tick: time.Hour, Rand: seeded()})
	addr := startServer(t, srv)
	login(t, addr, "Bertil")
	c := login(t, addr, "Anna")
	c.expectContains(fmt.Sprintf(MsgHere, "Bertil"+MsgListSep+"a black cat"))
}

func TestWandererMoves(t *testing.T) {
	srv := NewServerWithOptions(Options{World: npcWorld(t), Tick: 20 * time.Millisecond, Rand: seeded()})
	addr := startServer(t, srv)
	c := login(t, addr, "Anna")
	// With chance 1 the cat walks every tick: north, south, north, ...
	c.expectContains(fmt.Sprintf(MsgLeaves, "a black cat", directionTravel["north"]))
	c.expectContains(fmt.Sprintf(MsgArrives, "a black cat"))
}

func TestTalkerTalks(t *testing.T) {
	w := npcWorld(t)
	w.Start = "cell"
	srv := NewServerWithOptions(Options{World: w, Tick: 20 * time.Millisecond, Rand: seeded()})
	addr := startServer(t, srv)
	c := login(t, addr, "Anna")
	c.expectContains(fmt.Sprintf(MsgSays, "a parrot", "Pretty bird!"))
}

type countingBehavior struct{ n *atomic.Int64 }

func (b countingBehavior) Act(s *Server, n *NPC) { b.n.Add(1) }

func TestCustomBehavior(t *testing.T) {
	w := npcWorld(t)
	var count atomic.Int64
	w.NPCs["parrot"].AddBehavior(countingBehavior{&count})
	srv := NewServerWithOptions(Options{World: w, Tick: 2 * time.Millisecond, Rand: seeded()})
	startServer(t, srv)
	deadline := time.Now().Add(testTimeout)
	for count.Load() < 5 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if count.Load() < 5 {
		t.Errorf(L("ett eget Behavior ska köras varje tick (Act anropades %d gånger)", "a custom Behavior should run every tick (Act was called %d times)", "独自の Behavior は毎ティック実行されるべきです（Act の呼び出しは %d 回）"), count.Load())
	}
}

func TestNPCKinds(t *testing.T) {
	srv := NewServerWithOptions(Options{World: npcWorld(t), Tick: 20 * time.Millisecond, Rand: seeded()})
	addr := startServer(t, srv)
	c := jsonLoginKinds(t, addr, "Anna")
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		f := c.readFrame()
		text, _ := f["text"].(string)
		if strings.Contains(text, "a black cat") && f["type"] == "text" {
			if f["kind"] != KindMove {
				t.Errorf(L("när en varelse går ska kind vara %q, var %v", "when a creature moves, kind should be %q, was %v", "生き物が移動したとき、kind は %q であるべきですが、%v でした"), KindMove, f["kind"])
			}
			return
		}
	}
	t.Error(L("katten rörde sig aldrig", "the cat never moved", "猫が一度も動きませんでした"))
}

// jsonLoginKinds logs in with the JSON protocol and reads away the frames after login.
func jsonLoginKinds(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expect(Welcome)
	c.expect(MsgAskName)
	c.send("@hello mudschool 1")
	c.send(name)
	c.drain()
	return c
}
