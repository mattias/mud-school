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
		addr = args[0]
		if _, _, err := net.SplitHostPort(addr); err != nil {
			addr = net.JoinHostPort(addr, "4000") // "localhost" -> "localhost:4000"
		}
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
		"Ansluten till %s. Skriv och tryck Enter; Ctrl+C avslutar.",
		"Connected to %s. Type and press Enter; Ctrl+C quits.",
		"%s に接続しました。入力して Enter。Ctrl+C で終了します。"), addr)))

	// The keyboard -> the server, in its own goroutine.
	go func() {
		in := bufio.NewScanner(os.Stdin)
		for in.Scan() {
			if _, err := io.WriteString(conn, in.Text()+"\r\n"); err != nil {
				return
			}
		}
		// End of input (Ctrl+D, or Ctrl+Z Enter on Windows): stop sending.
		if tc, ok := conn.(*net.TCPConn); ok {
			tc.CloseWrite()
		}
	}()

	// The server -> the screen, line by line.
	out := bufio.NewReader(conn)
	for {
		line, err := out.ReadString('\n')
		if line != "" {
			fmt.Println(strings.TrimRight(line, "\r\n"))
		}
		if err != nil {
			break
		}
	}
	fmt.Println(dim(L("Servern stängde anslutningen.", "The server closed the connection.", "サーバーが接続を閉じました。")))
	return nil
}
