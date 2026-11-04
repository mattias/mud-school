package mud

import (
	"fmt"
	"testing"
	"time"
)

func testWorld() *World {
	return NewWorld("torg",
		&Room{ID: "torg", Name: "Square", Description: "A square.", Exits: map[string]string{"north": "krog", "east": "smedja"}},
		&Room{ID: "krog", Name: "Tavern", Description: "A tavern.\nIt smells of ale.", Exits: map[string]string{"south": "torg"}},
		&Room{ID: "smedja", Name: "Smithy", Description: "Hot.", Exits: map[string]string{"west": "torg"}},
		&Room{ID: "cell", Name: "Cell", Description: "No doors."},
	)
}

func exitsLine(dirs ...string) string {
	names := ""
	for i, d := range dirs {
		if i > 0 {
			names += MsgListSep
		}
		names += directionNames[d]
	}
	return fmt.Sprintf(MsgExits, names)
}

func TestNewWorld(t *testing.T) {
	w := testWorld()
	if w == nil {
		t.Fatal("NewWorld returned nil")
	}
	if w.Start != "torg" {
		t.Errorf("Start = %q, %s %q", w.Start, L("väntade", "expected", "期待値"), "torg")
	}
	if len(w.Rooms) != 4 {
		t.Errorf(L("världen har %d rum, väntade 4", "the world has %d rooms, expected 4", "部屋が %d 個あります。期待値は 4 です"), len(w.Rooms))
	}
	if r := w.Room("krog"); r == nil || r.Name != "Tavern" {
		t.Errorf("Room(%q) = %v, %s Tavern", "krog", r, L("väntade", "expected", "期待値"))
	}
	if r := w.Room("missing"); r != nil {
		t.Errorf(L("Room för okänt ID ska ge nil, fick %v", "Room for an unknown ID should return nil, got %v", "未知の ID に対して Room は nil を返すべきですが、%v でした"), r)
	}
}

func TestDefaultWorldIsConsistent(t *testing.T) {
	w := DefaultWorld()
	if w.Room(w.Start) == nil {
		t.Fatalf(L("startrummet %q finns inte", "the start room %q does not exist", "開始部屋 %q が存在しません"), w.Start)
	}
	if len(w.Rooms) < 5 {
		t.Errorf(L("standardvärlden har bara %d rum – minst 5, tack!", "the default world has only %d rooms – at least 5, please!", "標準の世界に部屋が %d 個しかありません。5 個以上にしてください！"), len(w.Rooms))
	}
	for id, r := range w.Rooms {
		if r.ID != id {
			t.Errorf(L("rummet %q ligger under nyckeln %q", "the room %q is stored under the key %q", "部屋 %q がキー %q に格納されています"), r.ID, id)
		}
		for dir, to := range r.Exits {
			if w.Room(to) == nil {
				t.Errorf(L("%s: utgången %s leder till %q som inte finns", "%s: exit %s leads to %q, which does not exist", "%s: 出口 %s の行き先 %q が存在しません"), id, dir, to)
			}
			if _, ok := directionNames[dir]; !ok {
				t.Errorf(L("%s: okänd riktning %q", "%s: unknown direction %q", "%s: 不明な方向 %q"), id, dir)
			}
		}
	}
}

func TestLookAfterLogin(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	c := login(t, addr, "Anna")
	c.expect("Square")
	c.expect("A square.")
	c.expect(exitsLine("north", "east"))

	c.send("look")
	c.expect("Square")
	c.expect("A square.")
	c.expect(exitsLine("north", "east"))
}

func TestMove(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	c := login(t, addr, "Anna")
	c.drain()
	c.send("north")
	c.expect("Tavern")
	c.expect("A tavern.")
	c.expect("It smells of ale.")
	c.expect(exitsLine("south"))

	c.send("w")
	c.expect(MsgCantGo)

	c.send("s")
	c.expect("Square")
	c.drain()
	c.send("e")
	c.expect("Smithy")
}

func TestOthersSeeYouComeAndGo(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.expectContains(fmt.Sprintf(MsgJoined, "Bertil"))
	bertil.expectContains(fmt.Sprintf(MsgHere, "Anna"))

	bertil.send("n")
	anna.expect(fmt.Sprintf(MsgLeaves, "Bertil", directionTravel["north"]))

	anna.send("n")
	bertil.expectContains(fmt.Sprintf(MsgArrives, "Anna"))
	anna.expectContains(fmt.Sprintf(MsgHere, "Bertil"))
}

func TestSayIsLocalShoutIsGlobal(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	cecilia := login(t, addr, "Cecilia")
	cecilia.send("e")
	cecilia.expectContains("Smithy")
	anna.drain()
	bertil.drain()
	cecilia.drain()

	anna.send("say hi")
	anna.expect(fmt.Sprintf(MsgYouSay, "hi"))
	bertil.expect(fmt.Sprintf(MsgSays, "Anna", "hi"))
	cecilia.expectSilence(200 * time.Millisecond)

	anna.send("shout HELLO")
	anna.expect(fmt.Sprintf(MsgYouShout, "HELLO"))
	bertil.expect(fmt.Sprintf(MsgShouts, "Anna", "HELLO"))
	cecilia.expect(fmt.Sprintf(MsgShouts, "Anna", "HELLO"))
}

func TestNewServerUsesDefaultWorld(t *testing.T) {
	addr := startServer(t, NewServer())
	c := login(t, addr, "Anna")
	w := DefaultWorld()
	c.expect(w.Room(w.Start).Name)
}
