package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// cmdConnect is a tiny line-based client, like telnet or nc, built into the
// course tool. Go writes Unicode correctly to every console (also on Windows,
// whatever the code page), so å, ä, ö and Japanese just work – no telnet,
// ncat or console settings needed.
func (c *course) cmdConnect(args []string) error {
	addr := "127.0.0.1:4000"
	if len(args) > 0 {
		addr = connectAddr(args[0])
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf(L(
			"kunde inte ansluta till %s: %v\nKör servern i en annan terminal: ./mudschool run server",
			"could not connect to %s: %v\nStart the server in another terminal: ./mudschool run server",
			"%s に接続できませんでした: %v\n別の端末でサーバーを起動してください: ./mudschool run server"), addr, err)
	}
	defer conn.Close()
	fmt.Println(dim(fmt.Sprintf(L(
		"Ansluten till %s. Skriv och tryck Enter. Avsluta med Ctrl+D (Windows: Ctrl+Z, Enter) eller Ctrl+C.",
		"Connected to %s. Type and press Enter. Quit with Ctrl+D (Windows: Ctrl+Z, Enter) or Ctrl+C.",
		"%s に接続しました。入力して Enter。終了は Ctrl+D（Windows: Ctrl+Z の後 Enter）か Ctrl+C。"), addr)))

	// The keyboard -> the server, in its own goroutine.
	go func() {
		in := bufio.NewScanner(os.Stdin)
		in.Buffer(make([]byte, 64*1024), 1024*1024)
		for in.Scan() {
			if _, err := io.WriteString(conn, in.Text()+"\r\n"); err != nil {
				return
			}
		}
		// End of input: stop sending, but keep showing what the server says.
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	// The server -> the screen, byte for byte as it arrives – also text
	// without a newline (a prompt, or a forgotten "\r\n" in lesson 01).
	var f telnetFilter
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			os.Stdout.Write(f.filter(buf[:n]))
		}
		if err != nil {
			break
		}
	}
	fmt.Println()
	fmt.Println(dim(L("Servern stängde anslutningen.", "The server closed the connection.", "サーバーが接続を閉じました。")))
	return nil
}

// connectAddr turns what the user typed into host:port (port 4000 if none):
// "localhost", "localhost:", "[::1]", "::1" and "10.0.0.5:4001" all work.
func connectAddr(s string) string {
	if host, port, err := net.SplitHostPort(s); err == nil {
		if port == "" {
			port = "4000"
		}
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(strings.Trim(s, "[]"), "4000")
}

// telnetFilter removes '\r' and telnet commands (IAC ...) from the stream,
// so that a server that negotiates like a real telnet server (for example to
// hide a password) doesn't print garbage. It keeps state between reads.
type telnetFilter struct {
	state int // 0 = text, 1 = after IAC, 2 = option byte, 3 = subnegotiation, 4 = IAC in subnegotiation
}

const iac = 255

func (f *telnetFilter) filter(in []byte) []byte {
	out := in[:0:0]
	for _, b := range in {
		switch f.state {
		case 0:
			switch b {
			case iac:
				f.state = 1
			case '\r':
			default:
				out = append(out, b)
			}
		case 1:
			switch {
			case b == iac: // IAC IAC = a literal 255
				out = append(out, b)
				f.state = 0
			case b >= 251 && b <= 254: // WILL, WONT, DO, DONT + an option
				f.state = 2
			case b == 250: // SB: subnegotiation until IAC SE
				f.state = 3
			default:
				f.state = 0
			}
		case 2:
			f.state = 0
		case 3:
			if b == iac {
				f.state = 4
			}
		case 4:
			if b == 240 { // SE
				f.state = 0
			} else {
				f.state = 3
			}
		}
	}
	return out
}
