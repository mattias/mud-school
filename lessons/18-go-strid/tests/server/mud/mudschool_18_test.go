package mud

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

// The world stands still (a tick of one hour) and the tests step it forward
// themselves with step(n), so that every fight is exactly reproducible.
//
// The tests don't depend on the exact random numbers either: a rat with 1 HP
// dies from any hit, and a troll that always does 30 damage beats a player
// with 20 HP in one round.
const combatWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "exits": {"north": "lair"}},
    {"id": "lair", "name": "Lair", "description": "It smells.", "terrain": "cave", "y": -1, "exits": {"south": "torg"}}
  ],
  "npcs": [
    {"id": "rat", "name": "a rat", "description": "Small.", "keywords": ["rat"], "room": "torg",
     "hp": 1, "damage": [1, 1], "attackable": true},
    {"id": "dove", "name": "a dove", "description": "Peaceful.", "keywords": ["dove"], "room": "torg"},
    {"id": "troll", "name": "a troll", "description": "Huge.", "keywords": ["troll"], "room": "lair",
     "hp": 1000, "damage": [30, 30], "attackable": true},
    {"id": "golem", "name": "a golem", "description": "Stone.", "keywords": ["golem"], "room": "lair",
     "hp": 1000, "damage": [1, 1], "attackable": true, "behaviors": [{"type": "wander", "chance": 1}]}
  ]
}`

type combatTest struct {
	srv  *Server
	addr string
}

func newCombatTest(t *testing.T) combatTest {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(combatWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, Rand: rand.New(rand.NewPCG(3, 4)), Respawn: 5})
	return combatTest{srv, startServer(t, srv)}
}

// step runs n ticks of the game loop, like the ticker would.
func (ct combatTest) step(n int) {
	for i := 0; i < n; i++ {
		ct.srv.call(func() {
			ct.srv.ticks++
			ct.srv.onTick()
		})
	}
}

// do sends a command and waits until the game loop has run it.
func (ct combatTest) do(c *testClient, cmd string) {
	c.send(cmd)
	c.drain()
}

func TestKillErrors(t *testing.T) {
	ct := newCombatTest(t)
	c := login(t, ct.addr, "Anna")
	c.drain()
	c.send("kill")
	c.expect(MsgAttackWhat)
	c.send("kill dragon")
	c.expect(fmt.Sprintf(MsgNoSuchTarget, "dragon"))
	c.send("kill dove")
	c.expect(fmt.Sprintf(MsgCantAttack, "a dove"))
	c.send("flee")
	c.expect(MsgNotFighting)
}

func TestDefeatAndRespawn(t *testing.T) {
	ct := newCombatTest(t)
	anna := login(t, ct.addr, "Anna")
	bertil := login(t, ct.addr, "Bertil")
	anna.drain()
	bertil.drain()

	anna.send("k rat")
	anna.expect(fmt.Sprintf(MsgYouAttack, "a rat"))
	bertil.expect(fmt.Sprintf(MsgAttacks, "Anna", "a rat"))
	ct.step(1)
	anna.expectContains(fmt.Sprintf(MsgYouKill, "a rat"))
	bertil.expectContains(fmt.Sprintf(MsgKills, "Anna", "a rat"))

	var inRoom bool
	ct.srv.call(func() {
		for _, n := range ct.srv.world.Room("torg").NPCs {
			if n.ID == "rat" {
				inRoom = true
			}
		}
	})
	if inRoom {
		t.Fatal(L("råttan är besegrad och ska inte finnas i rummet", "the rat is defeated and should not be in the room", "ネズミは倒されたので部屋にいないはずです"))
	}
	ct.step(5) // Respawn: 5
	anna.expectContains(fmt.Sprintf(MsgAppears, "a rat"))
	bertil.expectContains(fmt.Sprintf(MsgAppears, "a rat"))
}

func TestPlayerDies(t *testing.T) {
	ct := newCombatTest(t)
	c := login(t, ct.addr, "Anna")
	ct.do(c, "n")
	c.send("kill troll")
	c.expect(fmt.Sprintf(MsgYouAttack, "a troll"))
	ct.step(1)
	c.expectContains(fmt.Sprintf(MsgHitsYou, "a troll", 30))
	c.expectContains(fmt.Sprintf(MsgYouDie, "Square"))
	c.expectContains("Square") // the look after waking up
	c.drain()
	c.send("flee")
	c.expect(MsgNotFighting)
	var hp int
	ct.srv.call(func() { hp = ct.srv.players["Anna"].HP })
	if hp != PlayerMaxHP {
		t.Errorf(L("efter döden ska spelaren ha fullt hp (%d), har %d", "after dying the player should have full hp (%d), has %d", "倒れた後、プレイヤーの hp は満タン（%d）であるべきですが、%d でした"), PlayerMaxHP, hp)
	}
}

func TestStatusShowsHPAndTarget(t *testing.T) {
	ct := newCombatTest(t)
	c := jsonLoginKinds(t, ct.addr, "Anna")
	ct.do(c, "n")
	ct.do(c, "kill golem")
	ct.step(1)
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		f := c.readFrame()
		if text, _ := f["text"].(string); strings.Contains(text, "golem") && f["kind"] != KindCombat {
			t.Errorf(L("stridsmeddelanden ska ha kind %q: %v", "combat messages should have kind %q: %v", "戦闘メッセージの kind は %q であるべきです: %v"), KindCombat, f)
		}
		if f["type"] != "status" {
			continue
		}
		st := f["status"].(map[string]any)
		target, ok := st["target"].(map[string]any)
		if !ok {
			t.Fatalf(L("under striden ska status ha target: %v", "during the fight the status should have a target: %v", "戦闘中の status には target が必要です: %v"), st)
		}
		if target["name"] != "a golem" || target["max_hp"] != float64(1000) {
			t.Fatalf(L("status.target ska vara golemen, är %v", "status.target should be the golem, is %v", "status.target はゴーレムであるべきですが、%v でした"), target)
		}
		if hp, _ := target["hp"].(float64); hp >= 1000 {
			t.Fatalf(L("golemens hp ska minska när du slår den: %v", "the golem's hp should go down when you hit it: %v", "攻撃するとゴーレムの hp は減るべきです: %v"), target)
		}
		if st["max_hp"] != float64(PlayerMaxHP) || st["hp"] != float64(PlayerMaxHP-1) {
			t.Fatalf(L("status ska visa ditt hp efter en träff för 1 (%d/%d): %v", "status should show your hp after a hit for 1 (%d/%d): %v", "1 ダメージ受けた後の hp（%d/%d）が status に表示されるべきです: %v"), PlayerMaxHP-1, PlayerMaxHP, st)
		}
		return
	}
	t.Fatal(L("fick ingen status efter stridsrundan", "got no status after the combat round", "戦闘ラウンドの後に status が届きませんでした"))
}

func TestFlee(t *testing.T) {
	ct := newCombatTest(t)
	c := login(t, ct.addr, "Anna")
	ct.do(c, "n")
	ct.do(c, "kill troll")
	c.send("flee")
	c.expect(fmt.Sprintf(MsgYouFlee, directionTravel["south"]))
	c.expectContains("Square")
	ct.step(3)
	c.drain()
	var hp int
	var target *NPC
	ct.srv.call(func() { p := ct.srv.players["Anna"]; hp, target = p.HP, p.target })
	if target != nil || hp != PlayerMaxHP {
		t.Errorf(L("efter flykten ska striden vara slut (target %v, hp %d)", "after fleeing the fight should be over (target %v, hp %d)", "逃げた後は戦闘が終わっているべきです（target %v、hp %d）"), target, hp)
	}
}

func TestCreatureDoesNotWanderDuringFight(t *testing.T) {
	ct := newCombatTest(t)
	c := login(t, ct.addr, "Anna")
	ct.do(c, "n")
	ct.do(c, "kill golem")
	ct.step(10)
	var where string
	ct.srv.call(func() { where = ct.srv.world.NPCs["golem"].Location.ID })
	if where != "lair" {
		t.Errorf(L("golemen vandrade iväg mitt i striden (till %q)", "the golem wandered off in the middle of the fight (to %q)", "ゴーレムが戦闘中に歩き去りました（%q へ）"), where)
	}
	ct.do(c, "flee")
	ct.step(1)
	ct.srv.call(func() { where = ct.srv.world.NPCs["golem"].Location.ID })
	if where != "torg" {
		t.Errorf(L("när striden är slut ska golemen vandra igen (den står i %q)", "when the fight is over the golem should wander again (it is in %q)", "戦闘が終わったらゴーレムはまた歩き回るべきです（今は %q）"), where)
	}
}

func TestPlayerStartsWithFullHealth(t *testing.T) {
	ct := newCombatTest(t)
	login(t, ct.addr, "Anna")
	var hp, maxHP int
	ct.srv.call(func() {
		p := ct.srv.players["Anna"]
		hp, maxHP = p.HP, p.MaxHP
	})
	if hp != PlayerMaxHP || maxHP != PlayerMaxHP {
		t.Errorf(L("en ny spelare ska ha HP=MaxHP=%d, har %d/%d", "a new player should have HP=MaxHP=%d, has %d/%d", "新しいプレイヤーは HP=MaxHP=%d であるべきですが、%d/%d でした"), PlayerMaxHP, hp, maxHP)
	}
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
