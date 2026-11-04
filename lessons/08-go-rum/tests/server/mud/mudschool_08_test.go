package mud

import (
	"testing"
	"time"
)

func testWorld() *World {
	return NewWorld("torg",
		&Room{ID: "torg", Name: "Torget", Description: "Ett torg.", Exits: map[string]string{"north": "krog", "east": "smedja"}},
		&Room{ID: "krog", Name: "Krogen", Description: "En krog.\nDet luktar öl.", Exits: map[string]string{"south": "torg"}},
		&Room{ID: "smedja", Name: "Smedjan", Description: "Varmt.", Exits: map[string]string{"west": "torg"}},
		&Room{ID: "cell", Name: "Cellen", Description: "Inga dörrar."},
	)
}

func TestNewWorld(t *testing.T) {
	w := testWorld()
	if w.Start != "torg" {
		t.Errorf("Start = %q, väntade %q", w.Start, "torg")
	}
	if len(w.Rooms) != 4 {
		t.Errorf("världen har %d rum, väntade 4", len(w.Rooms))
	}
	if r := w.Room("krog"); r == nil || r.Name != "Krogen" {
		t.Errorf("Room(%q) = %v, väntade Krogen", "krog", r)
	}
	if r := w.Room("finns-inte"); r != nil {
		t.Errorf("Room för okänt ID ska ge nil, fick %v", r)
	}
}

func TestDefaultWorldIsConsistent(t *testing.T) {
	w := DefaultWorld()
	if w.Room(w.Start) == nil {
		t.Fatalf("startrummet %q finns inte", w.Start)
	}
	if len(w.Rooms) < 5 {
		t.Errorf("standardvärlden har bara %d rum – minst 5, tack!", len(w.Rooms))
	}
	for id, r := range w.Rooms {
		if r.ID != id {
			t.Errorf("rummet %q ligger under nyckeln %q", r.ID, id)
		}
		for dir, to := range r.Exits {
			if w.Room(to) == nil {
				t.Errorf("%s: utgången %s leder till %q som inte finns", id, dir, to)
			}
			if _, ok := directionNames[dir]; !ok {
				t.Errorf("%s: okänd riktning %q", id, dir)
			}
		}
	}
}

func TestLookAfterLogin(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	c := login(t, addr, "Anna")
	c.expect("Torget")
	c.expect("Ett torg.")
	c.expect("Utgångar: norr, öster")

	c.send("look")
	c.expect("Torget")
	c.expect("Ett torg.")
	c.expect("Utgångar: norr, öster")
}

func TestMove(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	c := login(t, addr, "Anna")
	c.drain()
	c.send("north")
	c.expect("Krogen")
	c.expect("En krog.")
	c.expect("Det luktar öl.")
	c.expect("Utgångar: söder")

	c.send("w")
	c.expect("Du kan inte gå åt det hållet.")

	c.send("s")
	c.expect("Torget")
	c.drain()
	c.send("e")
	c.expect("Smedjan")
}

func TestOthersSeeYouComeAndGo(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.expectContains("Bertil har anslutit.")
	bertil.expectContains("Här finns: Anna")

	bertil.send("n")
	anna.expect("Bertil går norrut.")

	anna.send("n")
	bertil.expectContains("Anna kommer in.")
	anna.expectContains("Här finns: Bertil")
}

func TestSayIsLocalShoutIsGlobal(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(testWorld()))
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	cecilia := login(t, addr, "Cecilia")
	cecilia.send("e")
	cecilia.expectContains("Smedjan")
	anna.drain()
	bertil.drain()
	cecilia.drain()

	anna.send("say hej")
	anna.expect("Du säger: hej")
	bertil.expect("Anna säger: hej")
	cecilia.expectSilence(200 * time.Millisecond)

	anna.send("shout HALLÅ")
	anna.expect("Du ropar: HALLÅ")
	bertil.expect("Anna ropar: HALLÅ")
	cecilia.expect("Anna ropar: HALLÅ")
}

func TestNewServerUsesDefaultWorld(t *testing.T) {
	addr := startServer(t, NewServer())
	c := login(t, addr, "Anna")
	w := DefaultWorld()
	c.expect(w.Room(w.Start).Name)
}
