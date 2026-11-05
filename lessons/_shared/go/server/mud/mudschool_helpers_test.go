// This file belongs to the course tool and is overwritten by `mudschool check`.
// It contains helpers that the lessons' tests use to talk to your server
// over TCP – exactly like a real client.
package mud

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

// L picks the test message in the course language ($MUDSCHOOL_LANG).
func L(sv, en, ja string) string {
	switch os.Getenv("MUDSCHOOL_LANG") {
	case "sv":
		return sv
	case "ja":
		return ja
	default:
		return en
	}
}

// startServer starts your server on a random free port and returns its address.
// If Serve panics, the panic is reported as a normal test failure with a hint.
func startServer(t *testing.T, srv *Server) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	var crashed atomic.Value
	go func() {
		defer func() {
			if r := recover(); r != nil {
				crashed.Store(fmt.Sprintf("%v\n%s", r, debug.Stack()))
			}
		}()
		srv.Serve(l)
	}()
	t.Cleanup(func() {
		l.Close()
		if msg, ok := crashed.Load().(string); ok {
			t.Errorf(L("Serve kraschade (panic): %s\nTips: anropar du net.Listen själv i Serve? Använd lyssnaren l som Serve får.",
				"Serve crashed (panic): %s\nHint: do you call net.Listen yourself in Serve? Use the listener l that Serve is given.",
				"Serve がクラッシュしました（panic）: %s\nヒント: Serve の中で自分で net.Listen を呼んでいませんか？ Serve が受け取ったリスナー l を使ってください。"), msg)
		}
	})
	return l.Addr().String()
}

type testClient struct {
	t    *testing.T
	name string
	conn net.Conn
	r    *bufio.Reader
	seen []string
}

func connect(t *testing.T, addr string) *testClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, testTimeout)
	if err != nil {
		t.Fatalf(L("kunde inte ansluta till servern: %v", "could not connect to the server: %v", "サーバーに接続できませんでした: %v"), err)
	}
	t.Cleanup(func() { conn.Close() })
	return &testClient{t: t, name: L("klient", "client", "クライアント"), conn: conn, r: bufio.NewReader(conn)}
}

func (c *testClient) send(line string) {
	c.t.Helper()
	c.conn.SetWriteDeadline(time.Now().Add(testTimeout))
	if _, err := io.WriteString(c.conn, line+"\r\n"); err != nil {
		c.t.Fatalf(L("%s: kunde inte skicka %q: %v", "%s: could not send %q: %v", "%s: %q を送信できませんでした: %v"), c.name, line, err)
	}
}

// readLine reads the next line. ok=false means the connection was closed.
func (c *testClient) readLine(timeout time.Duration) (line string, ok bool) {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	s, err := c.r.ReadString('\n')
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			hint := ""
			if len(c.seen) == 0 {
				hint = L("\nTips: servern har inte skickat något alls. Tar Serve emot anslutningar från lyssnaren l den får (och inte från en egen net.Listen)?",
					"\nHint: the server has not sent anything at all. Does Serve accept connections from the listener l it is given (and not from its own net.Listen)?",
					"\nヒント: サーバーからまだ何も届いていません。Serve は（自分で作った net.Listen ではなく）受け取ったリスナー l から接続を受け付けていますか？")
			}
			c.t.Fatalf(L("%s: väntade på en rad från servern men fick inget inom %v.\nRader hittills: %q%s",
				"%s: waited for a line from the server but got nothing within %v.\nLines so far: %q%s",
				"%s: サーバーからの行を待ちましたが %v 以内に何も届きませんでした。\nこれまでの行: %q%s"), c.name, timeout, c.seen, hint)
		}
		if s == "" {
			return "", false
		}
	}
	if strings.HasSuffix(s, "\n") && !strings.HasSuffix(s, "\r\n") {
		c.t.Errorf(L("%s: raden %q slutar med \\n men telnet förväntar sig \\r\\n",
			"%s: the line %q ends in \\n but telnet expects \\r\\n",
			"%s: 行 %q が \\n で終わっていますが、telnet では \\r\\n が必要です"), c.name, s)
	}
	s = strings.TrimRight(s, "\r\n")
	c.seen = append(c.seen, s)
	return s, true
}

// expect requires the next line to be exactly want.
func (c *testClient) expect(want string) {
	c.t.Helper()
	got, ok := c.readLine(testTimeout)
	if !ok {
		c.t.Fatalf(L("%s: väntade %q men servern stängde anslutningen",
			"%s: expected %q but the server closed the connection",
			"%s: %q を期待しましたが、サーバーが接続を閉じました"), c.name, want)
	}
	if got != want {
		c.t.Fatalf(L("%s: fel rad från servern\n  fick:     %q\n  väntade:  %q",
			"%s: wrong line from the server\n  got:      %q\n  expected: %q",
			"%s: サーバーからの行が違います\n  実際:  %q\n  期待:  %q"), c.name, got, want)
	}
}

// expectContains reads lines until one contains want, and returns it.
func (c *testClient) expectContains(want string) string {
	c.t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		got, ok := c.readLine(time.Until(deadline))
		if !ok {
			c.t.Fatalf(L("%s: väntade en rad som innehåller %q men servern stängde anslutningen.\nRader: %q",
				"%s: expected a line containing %q but the server closed the connection.\nLines: %q",
				"%s: %q を含む行を期待しましたが、サーバーが接続を閉じました。\n行: %q"), c.name, want, c.seen)
		}
		if strings.Contains(got, want) {
			return got
		}
	}
	c.t.Fatalf(L("%s: ingen rad innehöll %q.\nRader: %q", "%s: no line contained %q.\nLines: %q", "%s: %q を含む行がありませんでした。\n行: %q"), c.name, want, c.seen)
	return ""
}

// expectClosed requires the server to close the connection (lines before that are ignored).
func (c *testClient) expectClosed() {
	c.t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		if _, ok := c.readLine(time.Until(deadline)); !ok {
			return
		}
	}
	c.t.Fatalf(L("%s: väntade att servern skulle stänga anslutningen",
		"%s: expected the server to close the connection",
		"%s: サーバーが接続を閉じることを期待しました"), c.name)
}

// expectSilence requires the server to send nothing for a while.
func (c *testClient) expectSilence(d time.Duration) {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(d))
	s, err := c.r.ReadString('\n')
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() && s == "" {
		return
	}
	c.t.Fatalf(L("%s: väntade tystnad men fick %q (fel: %v)", "%s: expected silence but got %q (error: %v)", "%s: 何も届かないはずが %q を受信しました（エラー: %v）"), c.name, s, err)
}

// drain reads away everything the server has sent so far (until it has been
// quiet for a moment, or at most testTimeout if it never is).
func (c *testClient) drain() {
	c.t.Helper()
	stop := time.Now().Add(testTimeout)
	for time.Now().Before(stop) {
		c.conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
		if _, err := c.r.ReadString('\n'); err != nil {
			return
		}
	}
}

// login connects and logs in as name (from lesson 5).
func login(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expectContains(MsgAskName)
	c.send(name)
	c.expectContains(fmt.Sprintf(MsgWelcomeName, name))
	return c
}

// readFrame reads a line and decodes it as JSON.
func (c *testClient) readFrame() map[string]any {
	c.t.Helper()
	line, ok := c.readLine(testTimeout)
	if !ok {
		c.t.Fatalf(L("%s: väntade en JSON-rad men servern stängde anslutningen", "%s: expected a JSON line but the server closed the connection", "%s: JSON の行を期待しましたが、サーバーが接続を閉じました"), c.name)
	}
	var f map[string]any
	if err := json.Unmarshal([]byte(line), &f); err != nil {
		c.t.Fatalf(L("%s: raden %q är inte giltig JSON: %v", "%s: the line %q is not valid JSON: %v", "%s: 行 %q は正しい JSON ではありません: %v"), c.name, line, err)
	}
	if _, ok := f["type"].(string); !ok {
		c.t.Fatalf(L("%s: frame utan \"type\": %q", "%s: frame without \"type\": %q", "%s: \"type\" のないフレーム: %q"), c.name, line)
	}
	return f
}

// expectFrame reads frames until one of type typ arrives.
func (c *testClient) expectFrame(typ string) map[string]any {
	c.t.Helper()
	for i := 0; i < 20; i++ {
		if f := c.readFrame(); f["type"] == typ {
			return f
		}
	}
	c.t.Fatalf(L("%s: fick ingen frame av typen %q", "%s: got no frame of type %q", "%s: 種類 %q のフレームが届きませんでした"), c.name, typ)
	return nil
}
