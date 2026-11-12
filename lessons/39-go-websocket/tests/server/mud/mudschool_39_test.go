package mud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/coder/websocket"
)

var webTestFiles = fstest.MapFS{
	"index.html": {Data: []byte("<!doctype html><title>MUD</title>")},
	"mud.js":     {Data: []byte("// js")},
}

// newWebServer starts a MUD server with a gateway, both over HTTP (for the
// browser) and telnet. It returns the gateway's URL and the telnet address.
func newWebServer(t *testing.T) (*Server, string, string) {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(scriptWorldJSONNoNPC39))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	gw := NewGateway(webTestFiles)
	hs := httptest.NewServer(gw)
	go srv.Serve(gw.Listener())
	addr := startServer(t, srv)
	t.Cleanup(func() {
		shutdown(t, srv)
		hs.Close()
	})
	return srv, hs.URL, addr
}

const scriptWorldJSONNoNPC39 = `{
  "start": "square",
  "rooms": [{"id": "square", "name": "Square", "description": "A square.", "terrain": "town"}]
}`

// wsClient is a test player in a "browser".
type wsClient struct {
	t    *testing.T
	c    *websocket.Conn
	buf  string
	seen []string
}

func dialWeb(t *testing.T, url string, origin string) (*wsClient, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	opts := &websocket.DialOptions{HTTPHeader: http.Header{}}
	if origin != "" {
		opts.HTTPHeader.Set("Origin", origin)
	}
	c, _, err := websocket.Dial(ctx, strings.Replace(url, "http://", "ws://", 1)+"/ws", opts)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { c.CloseNow() })
	return &wsClient{t: t, c: c}, nil
}

func (w *wsClient) send(line string) {
	w.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	if err := w.c.Write(ctx, websocket.MessageText, []byte(line+"\n")); err != nil {
		w.t.Fatalf(L("webbklienten kunde inte skicka %q: %v", "the web client could not send %q: %v", "Web クライアントが %q を送れませんでした: %v"), line, err)
	}
}

// readLine returns the next line; ok is false when the connection closes.
func (w *wsClient) readLine() (string, bool) {
	w.t.Helper()
	for !strings.Contains(w.buf, "\n") {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		typ, data, err := w.c.Read(ctx)
		cancel()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				w.t.Fatalf(L("webbklienten väntade på en rad men fick inget. Rader hittills: %q", "the web client waited for a line but got nothing. Lines so far: %q", "Web クライアントは行を待ちましたが何も届きません。これまでの行: %q"), w.seen)
			}
			return "", false
		}
		if typ != websocket.MessageText {
			w.t.Errorf(L("servern ska skicka textmeddelanden, inte binära", "the server should send text messages, not binary ones", "サーバーはバイナリではなくテキストのメッセージを送るべきです"))
		}
		w.buf += string(data)
	}
	line, rest, _ := strings.Cut(w.buf, "\n")
	w.buf = rest
	line = strings.TrimRight(line, "\r")
	w.seen = append(w.seen, line)
	return line, true
}

func (w *wsClient) expectContains(want string) string {
	w.t.Helper()
	for i := 0; i < 50; i++ {
		line, ok := w.readLine()
		if !ok {
			w.t.Fatalf(L("webbklienten väntade %q men anslutningen stängdes. Rader: %q", "the web client expected %q but the connection closed. Lines: %q", "Web クライアントは %q を待ちましたが接続が閉じました。行: %q"), want, w.seen)
		}
		if strings.Contains(line, want) {
			return line
		}
	}
	w.t.Fatalf(L("webbklienten fick ingen rad med %q. Rader: %q", "the web client got no line containing %q. Lines: %q", "Web クライアントに %q を含む行が届きません。行: %q"), want, w.seen)
	return ""
}

func TestConnListener(t *testing.T) {
	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1234}
	l := NewConnListener(addr)
	if l.Addr() != addr {
		t.Errorf(L("Addr ska returnera adressen lyssnaren skapades med", "Addr should return the address the listener was created with", "Addr は作成時のアドレスを返すべきです"))
	}
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	go func() {
		if !l.Offer(a) {
			t.Errorf(L("Offer ska lyckas när någon anropar Accept", "Offer should succeed when someone calls Accept", "誰かが Accept を呼べば Offer は成功するべきです"))
		}
	}()
	got, err := l.Accept()
	if err != nil || got != a {
		t.Fatalf(L("Accept ska returnera anslutningen som Offer lämnade, fick %v, %v", "Accept should return the connection Offer handed over, got %v, %v", "Accept は Offer が渡した接続を返すべきです: %v, %v"), got, err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := l.Accept() // waits ...
		done <- err
	}()
	time.Sleep(20 * time.Millisecond)
	l.Close()
	select {
	case err := <-done:
		if !errors.Is(err, net.ErrClosed) {
			t.Errorf(L("efter Close ska Accept ge ett fel som är net.ErrClosed (errors.Is), fick %v", "after Close, Accept should return an error that is net.ErrClosed (errors.Is), got %v", "Close の後、Accept は net.ErrClosed である（errors.Is）エラーを返すべきですが %v"), err)
		}
	case <-time.After(testTimeout):
		t.Fatal(L("Close ska väcka en Accept som väntar", "Close should wake up an Accept that is waiting", "Close は待っている Accept を起こすべきです"))
	}
	if l.Offer(b) {
		t.Errorf(L("Offer ska returnera false när lyssnaren är stängd", "Offer should return false when the listener is closed", "閉じたリスナーでは Offer は false を返すべきです"))
	}
	l.Close() // twice is fine
}

func TestTheGatewayServesFiles(t *testing.T) {
	_, url, _ := newWebServer(t)
	res, err := http.Get(url + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 || !strings.Contains(string(body), "<title>MUD</title>") {
		t.Errorf(L("GET / ska ge index.html (200), fick %d %q", "GET / should give index.html (200), got %d %q", "GET / は index.html（200）を返すべきですが %d %q"), res.StatusCode, body)
	}
	res, err = http.Get(url + "/nothing.txt")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 404 {
		t.Errorf(L("en fil som inte finns ska ge 404, fick %d", "a file that doesn't exist should give 404, got %d", "存在しないファイルは 404 になるべきですが %d"), res.StatusCode)
	}
}

func TestPlayInTheBrowser(t *testing.T) {
	_, url, addr := newWebServer(t)
	w, err := dialWeb(t, url, "")
	if err != nil {
		t.Fatalf(L("kunde inte öppna en WebSocket mot /ws: %v", "could not open a WebSocket to /ws: %v", "/ws への WebSocket を開けませんでした: %v"), err)
	}
	w.expectContains(MsgAskName)
	w.send("Anna")
	w.expectContains(fmt.Sprintf(MsgWelcomeName, "Anna"))
	w.expectContains("Square")

	// A telnet player and a browser player meet.
	b := login(t, addr, "Bertil")
	b.drain()
	w.send("say hello from the web")
	b.expectContains(fmt.Sprintf(MsgSays, "Anna", "hello from the web"))
	b.send("who")
	b.expectContains("Anna")
}

func TestJSONInTheBrowser(t *testing.T) {
	_, url, _ := newWebServer(t)
	w, err := dialWeb(t, url, "")
	if err != nil {
		t.Fatal(err)
	}
	w.send(HelloPrefix + " mudschool 1")
	w.expectContains(`"type":"hello"`)
	w.send("Anna")
	w.expectContains(`"type":"zone"`)
}

func TestOtherSitesMayNotConnect(t *testing.T) {
	_, url, _ := newWebServer(t)
	if _, err := dialWeb(t, url, "https://evil.example"); err == nil {
		t.Errorf(L("en WebSocket från en annan webbplats (Origin: https://evil.example) ska nekas – annars kan vilken sida som helst logga in sina besökare i din MUD",
			"a WebSocket from another site (Origin: https://evil.example) should be refused – otherwise any page could log its visitors into your MUD",
			"他のサイトからの WebSocket（Origin: https://evil.example）は拒否するべきです。でないと、どんなページでも訪問者をあなたの MUD にログインさせられます"))
	}
	host := strings.TrimPrefix(url, "http://")
	if _, err := dialWeb(t, url, "http://"+host); err != nil {
		t.Errorf(L("en WebSocket från sidan själv ska tillåtas: %v", "a WebSocket from the page itself should be allowed: %v", "ページ自身からの WebSocket は許可するべきです: %v"), err)
	}
}

func TestShutdownReachesTheBrowser(t *testing.T) {
	w0, err := LoadWorld(strings.NewReader(scriptWorldJSONNoNPC39))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w0, Tick: time.Hour})
	gw := NewGateway(webTestFiles)
	hs := httptest.NewServer(gw)
	defer hs.Close()
	served := make(chan error, 1)
	go func() { served <- srv.Serve(gw.Listener()) }()
	w, err := dialWeb(t, hs.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	w.expectContains(MsgAskName)
	w.send("Anna")
	w.expectContains("Square")
	shutdown(t, srv)
	goodbye := false
	for {
		line, ok := w.readLine()
		if !ok {
			break // closed – good
		}
		goodbye = goodbye || strings.Contains(line, MsgShutdown)
	}
	if !goodbye {
		t.Errorf(L("webbspelaren ska få MsgShutdown innan anslutningen stängs – en läs-tidsgräns får inte stänga WebSocketen (newDeadlineConn). Rader: %q",
			"the web player should get MsgShutdown before the connection closes – a read deadline must not close the WebSocket (newDeadlineConn). Lines: %q",
			"Web プレイヤーは接続が閉じる前に MsgShutdown を受け取るべきです。読み込み期限で WebSocket を閉じてはいけません（newDeadlineConn）。行: %q"), w.seen)
	}
	select {
	case err := <-served:
		if !errors.Is(err, ErrServerClosed) {
			t.Errorf(L("Serve(gw.Listener()) ska returnera ErrServerClosed efter Shutdown, fick %v", "Serve(gw.Listener()) should return ErrServerClosed after Shutdown, got %v", "Shutdown の後、Serve(gw.Listener()) は ErrServerClosed を返すべきですが %v"), err)
		}
	case <-time.After(testTimeout):
		t.Errorf(L("Serve(gw.Listener()) returnerade aldrig efter Shutdown", "Serve(gw.Listener()) never returned after Shutdown", "Shutdown の後も Serve(gw.Listener()) が戻りません"))
	}
}

func TestBrowserPlayersHaveTheirAddress(t *testing.T) {
	gw := NewGateway(webTestFiles)
	hs := httptest.NewServer(gw)
	defer hs.Close()
	defer gw.Listener().Close()
	dialed := make(chan *wsClient, 1)
	go func() {
		w, _ := dialWeb(t, hs.URL, "")
		dialed <- w
	}()
	conn, err := gw.Listener().Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if w := <-dialed; w != nil {
		defer w.c.CloseNow() // runs before conn.Close: no waiting for a goodbye
	}
	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil || host != "127.0.0.1" {
		t.Errorf(L("RemoteAddr för en webbspelare ska vara webbläsarens adress (127.0.0.1:…), fick %v", "a web player's RemoteAddr should be the browser's address (127.0.0.1:…), got %v", "Web プレイヤーの RemoteAddr はブラウザのアドレス（127.0.0.1:…）であるべきですが %v"), conn.RemoteAddr())
	}
}
