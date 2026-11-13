package mud

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"testing/fstest"

	"github.com/coder/websocket"
)

func getHealth(t *testing.T, h http.Handler) (int, Health) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	var body Health
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf(L("/healthz ska svara med JSON: %q", "/healthz should answer with JSON: %q", "/healthz は JSON で答えるべきです: %q"), rec.Body.String())
	}
	return rec.Code, body
}

func TestHealth(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(oneRoomWorld42))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	addr := startServer(t, srv)
	h := srv.HealthHandler()
	code, body := getHealth(t, h)
	if code != 200 || body.Status != "ok" || body.Players != 0 {
		t.Errorf(L("en frisk server utan spelare: 200 {status: ok, players: 0}, fick %d %+v", "a healthy server without players: 200 {status: ok, players: 0}, got %d %+v", "プレイヤーのいない健康なサーバー: 200 {status: ok, players: 0} のはずが %d %+v"), code, body)
	}
	c := login(t, addr, "Anna")
	c.drain()
	if _, body := getHealth(t, h); body.Players != 1 {
		t.Errorf(L("players ska räkna inloggade spelare, fick %d", "players should count the players logged in, got %d", "players はログイン中のプレイヤー数のはずが %d"), body.Players)
	}
	shutdown(t, srv)
	if code, body := getHealth(t, h); code != http.StatusServiceUnavailable || body.Status != "stopping" {
		t.Errorf(L("efter Shutdown: 503 {status: stopping}, fick %d %+v", "after Shutdown: 503 {status: stopping}, got %d %+v", "Shutdown の後は 503 {status: stopping} のはずが %d %+v"), code, body)
	}
}

// acceptOne opens a WebSocket with the headers and returns the address the
// server sees.
func acceptOne(t *testing.T, gw *Gateway, header http.Header) string {
	t.Helper()
	hs := httptest.NewServer(gw)
	defer hs.Close()
	dialed := make(chan *websocket.Conn, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()
		c, _, err := websocket.Dial(ctx, strings.Replace(hs.URL, "http://", "ws://", 1)+"/ws", &websocket.DialOptions{HTTPHeader: header})
		if err != nil {
			t.Errorf("dial: %v", err)
		}
		dialed <- c
	}()
	conn, err := gw.Listener().Accept()
	if err != nil {
		t.Fatal(err)
	}
	if c := <-dialed; c != nil {
		c.CloseNow()
	}
	conn.Close()
	host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	return host
}

func TestTrustProxy(t *testing.T) {
	xff := http.Header{"X-Forwarded-For": {"203.0.113.7, 198.51.100.2"}}

	gw := NewGateway(webFiles42)
	if got := acceptOne(t, gw, xff); got != "127.0.0.1" {
		t.Errorf(L("utan TrustProxy ska X-Forwarded-For inte tros på – vem som helst kan skicka den. Adressen ska vara 127.0.0.1, fick %q",
			"without TrustProxy, X-Forwarded-For must not be believed – anyone can send it. The address should be 127.0.0.1, got %q",
			"TrustProxy なしでは X-Forwarded-For を信じてはいけません（誰でも送れます）。アドレスは 127.0.0.1 のはずが %q"), got)
	}
	gw = NewGateway(webFiles42)
	gw.TrustProxy = true
	if got := acceptOne(t, gw, xff); got != "198.51.100.2" {
		t.Errorf(L("med TrustProxy är spelarens adress den sista i X-Forwarded-For (den som proxyn själv lade till): väntade 198.51.100.2, fick %q",
			"with TrustProxy the player's address is the last one in X-Forwarded-For (the one the proxy added itself): expected 198.51.100.2, got %q",
			"TrustProxy ありなら、プレイヤーのアドレスは X-Forwarded-For の最後（プロキシ自身が足したもの）: 198.51.100.2 のはずが %q"), got)
	}
	gw = NewGateway(webFiles42)
	gw.TrustProxy = true
	if got := acceptOne(t, gw, http.Header{}); got != "127.0.0.1" {
		t.Errorf(L("med TrustProxy men utan rubriken: anslutningens egen adress, fick %q", "with TrustProxy but no header: the connection's own address, got %q", "TrustProxy ありでもヘッダーがなければ接続自身のアドレス: %q"), got)
	}
}

const oneRoomWorld42 = `{
  "start": "square",
  "rooms": [{"id": "square", "name": "Square", "description": "A square.", "terrain": "town"}]
}`

var webFiles42 = fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>MUD</title>")}}
