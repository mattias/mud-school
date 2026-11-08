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

const scriptWorldJSON = `{
  "start": "torg",
  "rooms": [{"id": "torg", "name": "Square", "description": "A square.", "terrain": "town"}],
  "npcs": [{"id": "dog", "name": "a dog", "description": "Friendly.", "keywords": ["dog"], "room": "torg",
            "behaviors": [{"type": "script", "script": "bark"}]}]
}`

const danceLua = `
mud.command("dance", function(player, args)
  mud.send(player, player .. " dances " .. args .. "!")
end)

mud.command("look", function(player, args)
  mud.send(player, "a script must not replace look")
end)

mud.command("probe", function(player, args)
  mud.send(player, tostring(os) .. "," .. tostring(io) .. "," .. tostring(dofile))
end)

mud.command("whoami", function(player, args)
  local p = mud.player(player)
  mud.send(player, p.name .. " in " .. p.room .. " with " .. p.hp .. "/" .. p.max_hp)
end)

mud.command("forever", function(player, args)
  while true do end
end)

mud.behavior("bark", function(npc, room)
  mud.broadcast(room, npc .. " says: Woof!")
end)
`

type scriptTest struct {
	srv  *Server
	addr string
	dir  string
}

func newScriptTest(t *testing.T, every time.Duration, files map[string]string) scriptTest {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w, err := LoadWorld(strings.NewReader(scriptWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, ScriptDir: dir, ReloadEvery: every,
		Admins: []string{"Anna"}, Rand: rand.New(rand.NewPCG(1, 2))})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return scriptTest{srv, addr, dir}
}

func TestScriptCommands(t *testing.T) {
	st := newScriptTest(t, 0, map[string]string{"dance.lua": danceLua})
	c := login(t, st.addr, "Anna")
	c.drain()
	c.send("dance wildly")
	c.expect("Anna dances wildly!")
	c.send("whoami")
	c.expect("Anna in torg with 20/20")
	c.send("look")
	c.expect("Square")
	c.drain()
}

func TestTheSandbox(t *testing.T) {
	st := newScriptTest(t, 0, map[string]string{"dance.lua": danceLua})
	c := login(t, st.addr, "Anna")
	c.drain()
	c.send("probe")
	c.expect("nil,nil,nil")
}

func TestScriptsHaveATimeLimit(t *testing.T) {
	st := newScriptTest(t, 0, map[string]string{"dance.lua": danceLua})
	c := login(t, st.addr, "Anna")
	c.drain()
	start := time.Now()
	c.send("forever")
	c.expect(MsgScriptError)
	if d := time.Since(start); d > time.Second {
		t.Errorf(L("en oändlig loop ska stoppas efter ScriptTimeout, tog %v", "an endless loop should be stopped after ScriptTimeout, took %v", "無限ループは ScriptTimeout の後に止めるべきですが %v かかりました"), d)
	}
	c.send("dance again")
	c.expect("Anna dances again!")
}

func TestScriptBehaviours(t *testing.T) {
	st := newScriptTest(t, 0, map[string]string{"dance.lua": danceLua})
	c := login(t, st.addr, "Anna")
	c.drain()
	st.srv.call(func() {
		st.srv.ticks++
		st.srv.onTick()
	})
	c.expect("dog says: Woof!")
}

func TestScriptsReload(t *testing.T) {
	st := newScriptTest(t, 20*time.Millisecond, map[string]string{"dance.lua": danceLua})
	c := login(t, st.addr, "Anna")
	c.drain()
	time.Sleep(50 * time.Millisecond)
	tango := `mud.command("dance", function(player, args) mud.send(player, player .. " dances a tango") end)`
	if err := os.WriteFile(filepath.Join(st.dir, "dance.lua"), []byte(tango), 0o644); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		c.send("dance")
		if line, _ := c.readLine(testTimeout); line == "Anna dances a tango" {
			return
		}
		time.Sleep(30 * time.Millisecond)
	}
	t.Fatal(L("efter att dance.lua ändrats ska det nya skriptet användas (utan omstart)", "after dance.lua has changed, the new script should be used (without a restart)", "dance.lua を変えたら、（再起動なしで）新しいスクリプトが使われるべきです"))
}

func TestBrokenScripts(t *testing.T) {
	st := newScriptTest(t, 20*time.Millisecond, map[string]string{"dance.lua": danceLua})
	anna := login(t, st.addr, "Anna")
	anna.drain()
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(st.dir, "broken.lua"), []byte("this is not lua"), 0o644); err != nil {
		t.Fatal(err)
	}
	line := anna.expectContains(strings.SplitN(MsgScriptLoadFailed, ":", 2)[0])
	if !strings.Contains(line, "broken.lua") {
		line, _ = anna.readLine(testTimeout)
	}
	if !strings.Contains(line, "broken.lua") {
		t.Errorf(L("felet ska säga vilken fil: %q", "the error should say which file: %q", "エラーにはファイル名を含めるべきです: %q"), line)
	}
	anna.drain()
	anna.send("dance on")
	anna.expect(fmt.Sprintf("%s dances on!", "Anna"))
}
