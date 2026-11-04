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
	for _, line := range []string{"hej", "titta på dig själv", "åäö fungerar också"} {
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
	second.send("igen")
	second.expect("igen")
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
		t.Fatalf("Serve returnerade direkt (%v) – den ska fortsätta ta emot anslutningar i en loop", err)
	default:
	}

	l.Close()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Serve returnerade nil när lyssnaren stängdes – den ska returnera felet från Accept")
		}
	case <-time.After(testTimeout):
		t.Fatal("Serve returnerade inte efter att lyssnaren stängts")
	}
}
