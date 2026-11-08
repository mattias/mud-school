package mud

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const buildWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "exits": {"east": "shop"}},
    {"id": "shop", "name": "Shop", "description": "Wares.", "terrain": "inside", "x": 1, "exits": {"west": "torg"}},
    {"id": "loft", "name": "Loft", "description": "Dusty.", "terrain": "inside", "y": 1}
  ]
}`

func newBuildServer(t *testing.T, every time.Duration) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "world.json")
	if err := os.WriteFile(path, []byte(buildWorldJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	w, err := LoadWorldFile(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, WorldPath: path, ReloadEvery: every, Admins: []string{"Anna"}})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return addr, path
}

// waitForFile waits until the world file contains want (saving happens outside the game loop).
func waitForFile(t *testing.T, path, want string) *World {
	t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil && strings.Contains(string(data), want) {
			w, err := LoadWorldFile(path)
			if err != nil {
				t.Fatalf(L("världsfilen som sparades går inte att läsa: %v", "the saved world file can't be loaded: %v", "保存したワールドファイルを読み込めません: %v"), err)
			}
			return w
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf(L("världsfilen innehåller aldrig %q", "the world file never contains %q", "ワールドファイルに %q が現れません"), want)
	return nil
}

func TestDigAndDescribe(t *testing.T) {
	addr, path := newBuildServer(t, 0)
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.drain()
	bertil.drain()

	anna.send("@dig north Dark cave")
	anna.expect(fmt.Sprintf(MsgDug, directionNames["north"], "Dark cave", "dark-cave"))
	anna.send("@dig north Again")
	anna.expect(fmt.Sprintf(MsgExitExists, directionNames["north"]))

	// Visible to everyone at once.
	bertil.send("go north")
	bertil.expect("Dark cave")
	bertil.expect(MsgNewRoomDesc)
	bertil.drain()

	anna.send("go north")
	anna.drain()
	bertil.drain() // "Anna arrives."
	anna.send("@desc Water drips from the ceiling.")
	anna.expect(MsgBuilt)
	anna.send("@name The Dark Cave")
	anna.expect(MsgBuilt)
	bertil.send("look")
	bertil.expect("The Dark Cave")
	bertil.expect("Water drips from the ceiling.")

	w := waitForFile(t, path, "The Dark Cave") // saved after the description
	cave := w.Room("dark-cave")
	if cave == nil || cave.Name != "The Dark Cave" || cave.Exits["south"] != "torg" || w.Room("torg").Exits["north"] != "dark-cave" {
		t.Errorf(L("den sparade världen ska ha rummet med namn, beskrivning och utgångar åt båda hållen: %+v", "the saved world should have the room with its name, description and exits both ways: %+v", "保存した世界には、名前、説明、双方向の出口を持つ部屋があるべきです: %+v"), cave)
	}
	if cave != nil && cave.Y != -1 {
		t.Errorf(L("rummet norrut ska ligga på y = -1, ligger på %d", "the room to the north should be at y = -1, it is at %d", "北の部屋は y = -1 にあるべきですが %d です"), cave.Y)
	}
}

func TestBuildingIsValidated(t *testing.T) {
	addr, _ := newBuildServer(t, 0)
	anna := login(t, addr, "Anna")
	anna.drain()
	anna.send("@dig south Cellar") // the loft is already at (0, 1)
	anna.expectContains(strings.SplitN(MsgBuildInvalid, "\n", 2)[0])
	anna.drain()
	anna.send("go south")
	anna.expect(MsgCantGo)
	anna.send("@exit north loft") // the loft isn't north of the square
	anna.expectContains(strings.SplitN(MsgBuildInvalid, "\n", 2)[0])
	anna.drain()
	anna.send("@exit east none")
	anna.expect(MsgBuilt)
	anna.send("go east")
	anna.expect(MsgCantGo)
	anna.send("@dig")
	anna.expect(MsgDigUsage)
}

func TestOnlyAdminsBuild(t *testing.T) {
	addr, _ := newBuildServer(t, 0)
	bertil := login(t, addr, "Bertil")
	bertil.drain()
	for _, cmd := range []string{"@dig north X", "@desc Y", "@name Z", "@exit east none"} {
		bertil.send(cmd)
		bertil.expect(MsgNotAdmin)
	}
}

func TestOwnSavesAreNotReloaded(t *testing.T) {
	addr, path := newBuildServer(t, 20*time.Millisecond)
	anna := login(t, addr, "Anna")
	anna.drain()
	time.Sleep(50 * time.Millisecond)
	anna.send("@desc A busy square.")
	anna.expect(MsgBuilt)
	waitForFile(t, path, "A busy square.")
	anna.expectSilence(300 * time.Millisecond) // no MsgReloaded: we wrote it ourselves
}

func TestNoWorldFileNoSave(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(buildWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, Admins: []string{"Anna"}})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	anna := login(t, addr, "Anna")
	anna.drain()
	anna.send("@desc Still a square.")
	anna.expect(MsgBuilt)
	anna.expect(MsgNotSaved)
}
