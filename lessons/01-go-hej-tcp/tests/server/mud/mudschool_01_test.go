package mud

import (
	"net"
	"testing"
	"time"
)

func TestWelcome(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
}

func TestEcho(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	for _, line := range []string{"hello", "åäö works too", "日本語も大丈夫"} {
		c.send(line)
		c.expect(line)
	}
}

func TestClientCanReconnect(t *testing.T) {
	addr := startServer(t, NewServer())
	first := connect(t, addr)
	first.expect(Welcome)
	first.conn.Close()

	second := connect(t, addr)
	second.expect(Welcome)
	second.send("again")
	second.expect("again")
}

func TestServeReturnsErrorWhenListenerCloses(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- NewServer().Serve(l) }()

	time.Sleep(50 * time.Millisecond)
	select {
	case err := <-done:
		t.Fatalf(L("Serve returnerade direkt (%v) – den ska fortsätta ta emot anslutningar i en loop",
			"Serve returned immediately (%v) – it should keep accepting connections in a loop",
			"Serve がすぐに戻りました (%v)。ループで接続を受け付け続ける必要があります"), err)
	default:
	}

	l.Close()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal(L("Serve returnerade nil när lyssnaren stängdes – den ska returnera felet från Accept",
				"Serve returned nil when the listener was closed – it should return the error from Accept",
				"リスナーを閉じたとき Serve が nil を返しました。Accept のエラーを返す必要があります"))
		}
	case <-time.After(testTimeout):
		t.Fatal(L("Serve returnerade inte efter att lyssnaren stängts",
			"Serve did not return after the listener was closed",
			"リスナーを閉じても Serve が戻りませんでした"))
	}
}
