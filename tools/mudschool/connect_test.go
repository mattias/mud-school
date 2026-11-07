package main

import "testing"

func TestConnectAddr(t *testing.T) {
	for in, want := range map[string]string{
		"localhost":     "localhost:4000",
		"localhost:":    "localhost:4000",
		"10.0.0.5:4001": "10.0.0.5:4001",
		"::1":           "[::1]:4000",
		"[::1]":         "[::1]:4000",
		"[::1]:4777":    "[::1]:4777",
		":4777":         ":4777",
	} {
		if got := connectAddr(in); got != want {
			t.Errorf("connectAddr(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTelnetFilter(t *testing.T) {
	var f telnetFilter
	// IAC WILL ECHO, text with \r\n, IAC IAC, a subnegotiation – split across reads.
	parts := [][]byte{
		{255, 251}, {1, 'h', 'i', '\r', '\n', 255}, {255, 'x', 255, 250, 24, 1, 255}, {240, 0xc3, 0xa5},
	}
	var got []byte
	for _, p := range parts {
		got = append(got, f.filter(p)...)
	}
	want := []byte{'h', 'i', '\n', 255, 'x', 0xc3, 0xa5}
	if string(got) != string(want) {
		t.Errorf("filter = %v, want %v", got, want)
	}
}
