package mud

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

const copyoverWorldJSON = `{
  "start": "torg",
  "rooms": [{"id": "torg", "name": "Square", "description": "A square.", "terrain": "town"}],
  "items": [{"id": "lamp", "name": "a lamp", "description": "Brass.", "keywords": ["lamp"], "room": "torg"}]
}`

func copyoverWorld(t *testing.T) *World {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(copyoverWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestTheStateIsJSON(t *testing.T) {
	in := CopyoverState{Ticks: 42, Players: []CopyoverPlayer{
		{Name: "Anna", Room: "torg", X: 3, Y: 4, HP: 17, Items: []string{"lamp"}, Structured: true, File: 1},
	}}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out CopyoverState
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(in) != fmt.Sprint(out) {
		t.Errorf(L("tillståndet ska överleva JSON:\n  in:  %+v\n  ut:  %+v", "the state should survive JSON:\n  in:  %+v\n  out: %+v", "状態は JSON を通っても変わらないべきです:\n  入力: %+v\n  出力: %+v"), in, out)
	}
	if !strings.Contains(string(data), `"structured":true`) || !strings.Contains(string(data), `"file":1`) {
		t.Errorf(L("JSON-namnen ska vara små bokstäver (structured, file): %s", "the JSON names should be lower case (structured, file): %s", "JSON の名前は小文字（structured、file）であるべきです: %s"), data)
	}
}

func TestCopyoverKeepsEveryoneConnected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("copyover needs file descriptors that can be passed on – try it in WSL")
	}
	old := NewServerWithOptions(Options{World: copyoverWorld(t), Tick: time.Hour})
	addr := startServer(t, old)
	anna := loginJSON(t, addr, "Anna")
	anna.expectZone("torg")
	bertil := login(t, addr, "Bertil")
	anna.drain()
	bertil.drain()
	anna.send("get lamp")
	anna.drain()
	anna.send("e")
	anna.drain()
	bertil.drain()

	state, files, err := old.PrepareCopyover()
	if err != nil {
		t.Fatalf("PrepareCopyover: %v", err)
	}
	if len(state.Players) != 2 || len(files) != 3 {
		t.Fatalf(L("två spelare och tre filer (lyssnaren + två anslutningar), fick %d och %d", "two players and three files (the listener + two connections), got %d and %d", "プレイヤー 2 人とファイル 3 つ（リスナー + 接続 2 つ）のはずが %d と %d でした"), len(state.Players), len(files))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := old.HandOver(ctx); err != nil {
		t.Fatalf("HandOver: %v", err)
	}

	// The new server – with new code, in real life – takes over.
	fresh := NewServerWithOptions(Options{World: copyoverWorld(t), Tick: time.Hour})
	l, err := fresh.Adopt(state, files)
	if err != nil {
		t.Fatalf("Adopt: %v", err)
	}
	go fresh.Serve(l)
	t.Cleanup(func() { shutdown(t, fresh) })

	bertil.expectContains(MsgCopyoverDone)
	if strings.Contains(strings.Join(bertil.seen, "\n"), fmt.Sprintf(MsgLeft, "Anna")) {
		t.Errorf(L("ingen ska ha loggats ut under copyover: %q", "nobody should have been logged out during the copyover: %q", "copyover の間に誰もログアウトしてはいけません: %q"), bertil.seen)
	}
	anna.expectContains(MsgCopyoverDone)
	anna.drain()
	anna.send("inventory")
	anna.expectContains("a lamp")
	anna.send("zone")
	z := anna.expectZone("torg")
	if e := entitiesOf(t, z["entities"])["p:Anna"]; e.x != state.Players[0].X || e.y != state.Players[0].Y {
		t.Errorf(L("Anna ska stå kvar där hon stod (%d,%d), står på (%d,%d)", "Anna should still stand where she stood (%d,%d), she is at (%d,%d)", "アンナは元の場所 (%d,%d) に立っているべきですが (%d,%d) にいます"), state.Players[0].X, state.Players[0].Y, e.x, e.y)
	}
	bertil.drain()
	bertil.send("who")
	bertil.expectContains("Anna")

	// And the listener goes on: new players can still come in.
	c := connect(t, addr)
	c.expectContains(MsgAskName)
}

func TestCopyoverNeedsAnAdmin(t *testing.T) {
	srv := NewServerWithOptions(Options{World: copyoverWorld(t), Tick: time.Hour, Admins: []string{"Anna"}})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	bertil := login(t, addr, "Bertil")
	bertil.drain()
	bertil.send("@copyover")
	bertil.expect(MsgNotAdmin)
	anna := login(t, addr, "Anna")
	anna.drain()
	anna.send("@copyover") // no OnCopyover: not available
	anna.expectContains(strings.SplitN(MsgCopyoverFailed, ":", 2)[0])
}
