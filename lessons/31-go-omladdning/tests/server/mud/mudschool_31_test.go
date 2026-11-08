package mud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const reloadWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "DESC", "terrain": "town", "exits": {"north": "hall"}},
    {"id": "hall", "name": "Hall", "description": "A hall.", "terrain": "inside", "y": -1, "exits": {"south": "torg"}}
  ],
  "items": [{"id": "lamp", "name": "a lamp", "description": "Brass.", "keywords": ["lamp"], "room": "torg"}]
}`

// writeWorld writes the world file the safe way: to a temporary file that is
// then renamed, so the server never reads half a file.
func writeWorld(t *testing.T, path, content string) {
	t.Helper()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}
}

func newReloadServer(t *testing.T, every time.Duration) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "world.json")
	writeWorld(t, path, strings.Replace(reloadWorldJSON, "DESC", "An old square.", 1))
	w, err := LoadWorldFile(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, WorldPath: path, ReloadEvery: every, Admins: []string{"anna"}})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return addr, path
}

func TestAtCommandsAreForAdmins(t *testing.T) {
	addr, _ := newReloadServer(t, 0)
	bertil := login(t, addr, "Bertil")
	bertil.drain()
	bertil.send("@reload")
	bertil.expect(MsgNotAdmin)
}

func TestAtReload(t *testing.T) {
	addr, path := newReloadServer(t, 0)
	anna := login(t, addr, "Anna") // "anna" in Admins: names are normalized
	anna.drain()
	writeWorld(t, path, strings.Replace(reloadWorldJSON, "DESC", "A brand new square.", 1))
	anna.send("@reload")
	anna.expectContains(MsgReloaded)
	anna.expectContains("A brand new square.")
}

func TestTheFileIsWatched(t *testing.T) {
	addr, path := newReloadServer(t, 20*time.Millisecond)
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.drain()
	bertil.drain()
	time.Sleep(50 * time.Millisecond)
	writeWorld(t, path, strings.Replace(reloadWorldJSON, "DESC", "The square, renovated.", 1))
	bertil.expectContains(MsgReloaded)
	bertil.expectContains("The square, renovated.")
}

func TestBrokenFileKeepsTheOldWorld(t *testing.T) {
	addr, path := newReloadServer(t, 20*time.Millisecond)
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.drain()
	bertil.drain()
	time.Sleep(50 * time.Millisecond)
	writeWorld(t, path, strings.Replace(reloadWorldJSON, `"exits": {"north": "hall"}`, `"exits": {"north": "nowhere"}`, 1))
	line := anna.expectContains(strings.SplitN(MsgReloadFailed, ":", 2)[0])
	if !strings.Contains(line, "nowhere") {
		line, _ = anna.readLine(testTimeout) // the error is on the next line
	}
	if !strings.Contains(line, "nowhere") {
		t.Errorf(L("felmeddelandet till administratören ska säga vad som är fel: %q", "the error message to the administrator should say what is wrong: %q", "管理者へのエラーメッセージは何が悪いかを伝えるべきです: %q"), line)
	}
	bertil.expectSilence(200 * time.Millisecond)
	bertil.send("look")
	bertil.expectContains("An old square.")
}

func TestPlayersKeepTheirThings(t *testing.T) {
	addr, path := newReloadServer(t, 0)
	anna := login(t, addr, "Anna")
	anna.drain()
	anna.send("get lamp")
	anna.drain()
	anna.send("go north")
	anna.drain()
	// The hall disappears from the world – and the lamp gets a new description.
	gone := `{
  "start": "torg",
  "rooms": [{"id": "torg", "name": "Square", "description": "Alone now.", "terrain": "town"}],
  "items": [{"id": "lamp", "name": "a shiny lamp", "description": "Polished.", "keywords": ["lamp"], "room": "torg"}]
}`
	writeWorld(t, path, gone)
	anna.send("@reload")
	anna.expectContains(MsgRoomGone)
	anna.expectContains("Alone now.")
	anna.drain()
	anna.send("inventory")
	anna.expectContains("a shiny lamp")
	anna.send("look")
	anna.expectContains("Alone now.")
	anna.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	for {
		line, err := anna.r.ReadString('\n')
		if strings.Contains(line, "lamp") {
			t.Fatalf(L("lampan ska vara i Annas ryggsäck – inte också på marken: %q", "the lamp should be in Anna's inventory – not on the ground as well: %q", "ランプはアンナの持ち物にあるべきで、地面にもあってはいけません: %q"), line)
		}
		if err != nil {
			break
		}
	}
}
