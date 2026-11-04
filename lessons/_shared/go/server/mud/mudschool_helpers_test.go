// Den här filen hör till kursverktyget och skrivs över av `mudschool check`.
// Den innehåller hjälpfunktioner som lektionernas tester använder för att
// prata med din server över TCP – precis som en riktig klient.
package mud

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

const testTimeout = 2 * time.Second

// startServer startar din server på en slumpad ledig port och returnerar adressen.
func startServer(t *testing.T, srv *Server) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("kunde inte lyssna: %v", err)
	}
	t.Cleanup(func() { l.Close() })
	go srv.Serve(l)
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
		t.Fatalf("kunde inte ansluta till servern: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &testClient{t: t, name: "klient", conn: conn, r: bufio.NewReader(conn)}
}

func (c *testClient) send(line string) {
	c.t.Helper()
	c.conn.SetWriteDeadline(time.Now().Add(testTimeout))
	if _, err := io.WriteString(c.conn, line+"\r\n"); err != nil {
		c.t.Fatalf("%s: kunde inte skicka %q: %v", c.name, line, err)
	}
}

// readLine läser nästa rad. ok=false betyder att anslutningen stängdes.
func (c *testClient) readLine(timeout time.Duration) (line string, ok bool) {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	s, err := c.r.ReadString('\n')
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			c.t.Fatalf("%s: väntade på en rad från servern men fick inget inom %v.\nRader hittills: %q", c.name, timeout, c.seen)
		}
		if s == "" {
			return "", false
		}
	}
	if strings.HasSuffix(s, "\n") && !strings.HasSuffix(s, "\r\n") {
		c.t.Errorf("%s: raden %q slutar med \\n men telnet förväntar sig \\r\\n", c.name, s)
	}
	s = strings.TrimRight(s, "\r\n")
	c.seen = append(c.seen, s)
	return s, true
}

// expect kräver att nästa rad är exakt want.
func (c *testClient) expect(want string) {
	c.t.Helper()
	got, ok := c.readLine(testTimeout)
	if !ok {
		c.t.Fatalf("%s: väntade %q men servern stängde anslutningen", c.name, want)
	}
	if got != want {
		c.t.Fatalf("%s: fel rad från servern\n  fick:     %q\n  väntade:  %q", c.name, got, want)
	}
}

// expectContains läser rader tills en rad innehåller want och returnerar den.
func (c *testClient) expectContains(want string) string {
	c.t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		got, ok := c.readLine(time.Until(deadline))
		if !ok {
			c.t.Fatalf("%s: väntade en rad som innehåller %q men servern stängde anslutningen.\nRader: %q", c.name, want, c.seen)
		}
		if strings.Contains(got, want) {
			return got
		}
	}
	c.t.Fatalf("%s: ingen rad innehöll %q.\nRader: %q", c.name, want, c.seen)
	return ""
}

// expectClosed kräver att servern stänger anslutningen (eventuella rader innan ignoreras).
func (c *testClient) expectClosed() {
	c.t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		if _, ok := c.readLine(time.Until(deadline)); !ok {
			return
		}
	}
	c.t.Fatalf("%s: väntade att servern skulle stänga anslutningen", c.name)
}

// expectSilence kräver att servern inte skickar något på en stund.
func (c *testClient) expectSilence(d time.Duration) {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(d))
	s, err := c.r.ReadString('\n')
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() && s == "" {
		return
	}
	c.t.Fatalf("%s: väntade tystnad men fick %q (fel: %v)", c.name, s, err)
}

// drain läser bort allt servern skickat fram till nu.
func (c *testClient) drain() {
	c.t.Helper()
	for {
		c.conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
		if _, err := c.r.ReadString('\n'); err != nil {
			return
		}
	}
}

// login ansluter och loggar in som name (från lektion 05).
func login(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expectContains("Vad heter du?")
	c.send(name)
	c.expectContains("Välkommen, " + name + "!")
	return c
}
