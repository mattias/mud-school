package mud

import (
	"expvar"
	"strings"
	"testing"
	"time"
)

func newProtectedServer(t *testing.T, o Options) string {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(oneRoomWorld43))
	if err != nil {
		t.Fatal(err)
	}
	o.World, o.Tick = w, time.Hour
	srv := NewServerWithOptions(o)
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return addr
}

func TestTheTokenBucket(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	r := NewRateLimiter(2, 3) // 2 per second, bursts of 3
	for i := 0; i < 3; i++ {
		if !r.Allow(t0) {
			t.Fatalf(L("en full hink släpper igenom burst (3) kommandon direkt – nummer %d nekades", "a full bucket lets burst (3) commands through at once – number %d was refused", "満杯のバケツは burst（3）個のコマンドをすぐ通します。%d 個目が拒否されました"), i+1)
		}
	}
	if r.Allow(t0) {
		t.Errorf(L("det fjärde kommandot i samma ögonblick ska nekas", "the fourth command at the same moment should be refused", "同じ瞬間の 4 個目は拒否されるべきです"))
	}
	if !r.Allow(t0.Add(500 * time.Millisecond)) {
		t.Errorf(L("efter 0,5 s har hinken fyllts på med 1 token (2 per sekund)", "after 0.5 s the bucket has refilled 1 token (2 per second)", "0.5 秒後には 1 トークン補充されています（毎秒 2）"))
	}
	if r.Allow(t0.Add(500 * time.Millisecond)) {
		t.Errorf(L("... men bara 1", "... but only 1", "……でも 1 つだけです"))
	}
	later := t0.Add(time.Hour)
	n := 0
	for r.Allow(later) {
		n++
	}
	if n != 3 {
		t.Errorf(L("efter en lång paus är hinken full igen – men aldrig mer än burst: %d kommandon gick igenom, väntade 3", "after a long pause the bucket is full again – but never more than burst: %d commands got through, expected 3", "長い休みの後バケツは満杯ですが burst を超えません: %d 個通りました（期待 3）"), n)
	}
}

func TestSpamIsSlowedDown(t *testing.T) {
	addr := newProtectedServer(t, Options{CommandRate: 5, CommandBurst: 5})
	c := login(t, addr, "Anna")
	c.drain()
	for i := 0; i < 30; i++ {
		c.send("say spam")
	}
	time.Sleep(1100 * time.Millisecond) // the bucket fills up again
	c.send("echo end-of-spam")
	c.expectContains("end-of-spam")
	slow, said := 0, 0
	for _, line := range c.seen {
		if strings.Contains(line, MsgSlowDown) {
			slow++
		}
		if strings.Contains(line, "spam") {
			said++
		}
	}
	if slow != 1 {
		t.Errorf(L("MsgSlowDown ska skickas en gång per flod, inte för varje rad: %d gånger", "MsgSlowDown should be sent once per flood, not for every line: %d times", "MsgSlowDown は洪水ごとに 1 回だけ送るべきです（行ごとではなく）: %d 回"), slow)
	}
	if said < 5 || said > 8 {
		t.Errorf(L("ungefär burst (5) kommandon ska gå igenom, %d gjorde det", "about burst (5) commands should get through, %d did", "およそ burst（5）個のコマンドが通るはずですが %d 個でした"), said)
	}
	c.send("look")
	c.expectContains("Square")
}

func TestFloodersAreDisconnected(t *testing.T) {
	addr := newProtectedServer(t, Options{CommandRate: 1, CommandBurst: 1})
	c := login(t, addr, "Anna")
	c.drain()
	var b strings.Builder
	for i := 0; i < FloodLimit+20; i++ {
		b.WriteString("look\r\n")
	}
	c.conn.Write([]byte(b.String()))
	c.expectContains(MsgFlooding)
	c.expectClosed()
}

func TestTooManyConnections(t *testing.T) {
	addr := newProtectedServer(t, Options{MaxPerIP: 2})
	a := connect(t, addr)
	a.expectContains(MsgAskName)
	b := connect(t, addr)
	b.expectContains(MsgAskName)
	c := connect(t, addr)
	c.expectContains(MsgTooManyConnections)
	c.expectClosed()

	a.conn.Close()
	deadline := time.Now().Add(testTimeout)
	for {
		d := connect(t, addr)
		line, ok := d.readLine(testTimeout)
		if ok && !strings.Contains(line, MsgTooManyConnections) {
			return // a place became free
		}
		if time.Now().After(deadline) {
			t.Fatal(L("när en anslutning stängs ska en ny få plats", "when a connection closes, a new one should get its place", "接続が閉じたら、新しい接続が入れるべきです"))
		}
		d.conn.Close()
		time.Sleep(50 * time.Millisecond)
	}
}

func TestLongLines(t *testing.T) {
	addr := newProtectedServer(t, Options{MaxLineLength: 100})
	c := login(t, addr, "Anna")
	c.drain()
	c.send("say " + strings.Repeat("x", 300))
	c.expectContains(MsgLineTooLong)
	c.expectClosed()
}

func TestLoginTimeout(t *testing.T) {
	addr := newProtectedServer(t, Options{LoginTimeout: 300 * time.Millisecond})
	c := connect(t, addr)
	c.expectContains(MsgAskName)
	start := time.Now()
	c.expectContains(MsgLoginTimeout)
	c.expectClosed()
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf(L("anslutningen ska stängas efter LoginTimeout, tog %v", "the connection should close after LoginTimeout, took %v", "接続は LoginTimeout の後に閉じるべきですが %v かかりました"), d)
	}

	// Once logged in, there is no time limit.
	p := login(t, addr, "Anna")
	time.Sleep(600 * time.Millisecond)
	p.drain()
	p.send("look")
	p.expectContains("Square")
}

func TestMetrics(t *testing.T) {
	get := func(name string) int64 {
		v, ok := expvar.Get(name).(*expvar.Int)
		if !ok {
			t.Fatalf(L("expvar %q saknas – skapa den med expvar.NewInt", "expvar %q is missing – create it with expvar.NewInt", "expvar %q がありません。expvar.NewInt で作りましょう"), name)
		}
		return v.Value()
	}
	conns, cmds := get("mud_connections_total"), get("mud_commands_total")
	addr := newProtectedServer(t, Options{})
	c := login(t, addr, "Anna")
	c.drain()
	c.send("look")
	c.send("who")
	c.drain()
	if get("mud_connections_total") < conns+1 || get("mud_commands_total") < cmds+2 {
		t.Errorf(L("mud_connections_total och mud_commands_total ska räknas upp", "mud_connections_total and mud_commands_total should count up", "mud_connections_total と mud_commands_total は数え上げるべきです"))
	}
	get("mud_players")
	get("mud_commands_dropped_total")
	get("mud_connections_refused_total")
}

const oneRoomWorld43 = `{
  "start": "square",
  "rooms": [{"id": "square", "name": "Square", "description": "A square.", "terrain": "town"}]
}`
