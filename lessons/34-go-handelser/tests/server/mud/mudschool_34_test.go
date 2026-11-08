package mud

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const eventWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "exits": {"north": "vault", "east": "shop"}},
    {"id": "vault", "name": "Vault", "description": "Gold!", "terrain": "inside", "y": -1, "exits": {"south": "torg"}},
    {"id": "shop", "name": "Shop", "description": "Wares.", "terrain": "inside", "x": 1, "exits": {"west": "torg"}}
  ],
  "npcs": [{"id": "rat", "name": "a rat", "description": "Small.", "keywords": ["rat"], "room": "torg",
            "hp": 1, "damage": [1, 1], "attackable": true}]
}`

type eventTest struct {
	srv  *Server
	addr string
}

// newEventTest creates the server; setup runs before it starts (to add hooks).
func newEventTest(t *testing.T, scripts map[string]string, setup func(s *Server)) eventTest {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(eventWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	o := Options{World: w, Tick: time.Hour, Rand: rand.New(rand.NewPCG(5, 6))}
	if scripts != nil {
		dir := t.TempDir()
		for name, content := range scripts {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		o.ScriptDir = dir
	}
	srv := NewServerWithOptions(o)
	if setup != nil {
		setup(srv)
	}
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return eventTest{srv, addr}
}

func (et eventTest) tick() {
	et.srv.call(func() {
		et.srv.ticks++
		et.srv.onTick()
	})
}

func TestAGoHookLocksADoor(t *testing.T) {
	et := newEventTest(t, nil, func(s *Server) {
		s.On(EventEnter, func(e *Event) {
			if e.To.ID == "vault" && e.Player.Name != "Banker" {
				e.Cancel("The vault door is locked.")
			}
		})
	})
	c := login(t, et.addr, "Anna")
	c.drain()
	c.send("go north")
	c.expect("The vault door is locked.")
	c.send("look")
	c.expect("Square")
	c.drain()
	c.send("go east")
	c.expect("Shop")
	c.drain()
	banker := login(t, et.addr, "Banker")
	banker.drain()
	banker.send("go north")
	banker.expect("Vault")
}

func TestHooksCanChangeWhatIsSaid(t *testing.T) {
	et := newEventTest(t, nil, func(s *Server) {
		s.On(EventSay, func(e *Event) {
			e.Text = strings.ReplaceAll(e.Text, "darn", "d**n")
		})
		s.On(EventSay, func(e *Event) {
			if strings.Contains(e.Text, "spam") {
				e.Cancel("No spam, please.")
			}
		})
	})
	anna := login(t, et.addr, "Anna")
	bertil := login(t, et.addr, "Bertil")
	anna.drain()
	bertil.drain()
	anna.send("say darn rats")
	anna.expect(fmt.Sprintf(MsgYouSay, "d**n rats"))
	bertil.expect(fmt.Sprintf(MsgSays, "Anna", "d**n rats"))
	anna.send("say buy spam")
	anna.expect("No spam, please.")
	bertil.expectSilence(200 * time.Millisecond)
}

func TestTickAndDeathEvents(t *testing.T) {
	ticks, deaths := 0, []string{}
	et := newEventTest(t, nil, func(s *Server) {
		s.On(EventTick, func(e *Event) { ticks++ })
		s.On(EventDeath, func(e *Event) {
			if e.NPC != nil {
				deaths = append(deaths, e.NPC.ID+" by "+e.Player.Name+" in "+e.Room.ID)
			}
		})
	})
	c := login(t, et.addr, "Anna")
	c.drain()
	c.send("kill rat")
	c.drain()
	et.tick()
	et.tick()
	var got int
	var died []string
	et.srv.call(func() { got, died = ticks, append([]string(nil), deaths...) })
	if got != 2 {
		t.Errorf(L("EventTick ska komma en gång per tick: %d på 2 tick", "EventTick should come once per tick: %d in 2 ticks", "EventTick はティックごとに 1 回: 2 ティックで %d 回"), got)
	}
	if fmt.Sprint(died) != "[rat by Anna in torg]" {
		t.Errorf(L("EventDeath när råttan dör (NPC, Player = den som dödade, Room): %v", "EventDeath when the rat dies (NPC, Player = the killer, Room): %v", "ネズミが死んだら EventDeath（NPC、Player = 倒した人、Room）: %v"), died)
	}
}

func TestLuaHooks(t *testing.T) {
	lua := `
mud.on("enter", function(ev)
  if ev.to == "vault" then
    return "A voice says: not today, " .. ev.player .. "."
  end
end)
mud.on("say", function(ev)
  if ev.text == "open sesame" then return "Nothing happens." end
end)
`
	et := newEventTest(t, map[string]string{"guard.lua": lua}, nil)
	c := login(t, et.addr, "Anna")
	c.drain()
	c.send("go north")
	c.expect("A voice says: not today, Anna.")
	c.send("say open sesame")
	c.expect("Nothing happens.")
	c.send("say hello")
	c.expect(fmt.Sprintf(MsgYouSay, "hello"))
	c.send("go east")
	c.expect("Shop")
}

func TestCancelWithoutAReason(t *testing.T) {
	et := newEventTest(t, nil, func(s *Server) {
		s.On(EventEnter, func(e *Event) { e.Cancel("") })
	})
	c := login(t, et.addr, "Anna")
	c.drain()
	c.send("go north")
	c.expect(MsgCancelled)
}
