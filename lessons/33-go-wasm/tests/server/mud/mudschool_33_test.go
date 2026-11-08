package mud

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// greeter.wasm is the plugin in plugins/greeter, built with
// cargo build --release --target wasm32-unknown-unknown.

// oldABI is a tiny hand-written WebAssembly module with every export the
// server wants – but abi_version() returns 2.
var oldABI = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x15, 0x04, 0x60, 0x00, 0x01, 0x7f, 0x60, 0x01, 0x7f, 0x01, 0x7f, 0x60, 0x02, 0x7f, 0x7f, 0x00, 0x60, 0x02, 0x7f, 0x7f, 0x01, 0x7f, 0x03, 0x05, 0x04, 0x00, 0x01, 0x02, 0x03, 0x05, 0x03, 0x01, 0x00, 0x01, 0x07, 0x37, 0x05, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, 0x0b, 0x61, 0x62, 0x69, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x00, 0x00, 0x05, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x00, 0x01, 0x07, 0x64, 0x65, 0x61, 0x6c, 0x6c, 0x6f, 0x63, 0x00, 0x02, 0x0a, 0x6f, 0x6e, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x00, 0x03, 0x0a, 0x13, 0x04, 0x04, 0x00, 0x41, 0x02, 0x0b, 0x04, 0x00, 0x41, 0x00, 0x0b, 0x02, 0x00, 0x0b, 0x04, 0x00, 0x41, 0x00, 0x0b}

func pluginDir(t *testing.T, extra map[string][]byte) string {
	t.Helper()
	dir := t.TempDir()
	greeter, err := os.ReadFile(filepath.Join("mudschool_testdata", "greeter.wasm"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"greeter.wasm": greeter}
	for k, v := range extra {
		files[k] = v
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func newPluginServer(t *testing.T, dir string) string {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(scriptWorldJSONNoNPC))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, PluginDir: dir, Admins: []string{"Anna"}})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return addr
}

const scriptWorldJSONNoNPC = `{
  "start": "torg",
  "rooms": [{"id": "torg", "name": "Square", "description": "A square.", "terrain": "town"}]
}`

func TestARustPlugin(t *testing.T) {
	c := login(t, newPluginServer(t, pluginDir(t, nil)), "Anna")
	c.drain()
	c.send("greet")
	c.expect("Hello, Anna! Greetings from a plugin written in Rust.")
	c.send("reverse åäö MUD")
	c.expect("DUM öäå")
	c.send("dance")
	c.expect(fmt.Sprintf(MsgUnknownCommand, "dance"))
}

func TestPluginsHaveATimeLimit(t *testing.T) {
	c := login(t, newPluginServer(t, pluginDir(t, nil)), "Bertil")
	c.drain()
	start := time.Now()
	c.send("forever")
	c.expect(MsgPluginError)
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf(L("ett plugin som aldrig blir klart ska stoppas efter PluginTimeout, tog %v", "a plugin that never finishes should be stopped after PluginTimeout, took %v", "終わらないプラグインは PluginTimeout の後に止めるべきですが %v かかりました"), d)
	}
	c.send("look")
	c.expect("Square")
}

func TestBadPlugins(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(scriptWorldJSONNoNPC))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	dir := pluginDir(t, map[string][]byte{"aaa-old.wasm": oldABI, "bbb-junk.wasm": []byte("not wasm at all")})
	pl, err := srv.LoadPlugins(dir)
	if pl != nil {
		defer pl.Close()
	}
	if err == nil {
		t.Fatal(L("LoadPlugins ska rapportera trasiga plugins", "LoadPlugins should report broken plugins", "LoadPlugins は壊れたプラグインを報告するべきです"))
	}
	msg := err.Error()
	for _, want := range []string{"aaa-old.wasm", "bbb-junk.wasm"} {
		if !strings.Contains(msg, want) {
			t.Errorf(L("felet ska nämna %s: %v", "the error should mention %s: %v", "エラーには %s を含めるべきです: %v"), want, msg)
		}
	}
	if !strings.Contains(msg, "2") {
		t.Errorf(L("ett plugin med fel ABI-version ska avvisas och säga vilken version det har: %v", "a plugin with the wrong ABI version should be rejected, saying which version it has: %v", "ABI バージョンの違うプラグインは、そのバージョンを示して拒否するべきです: %v"), msg)
	}
	if pl == nil || len(pl.list) != 1 || pl.list[0].Name != "greeter" {
		t.Errorf(L("de fungerande pluginen ska laddas ändå", "the working plugins should still load", "動くプラグインはそれでも読み込むべきです"))
	}
}
