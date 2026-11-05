package mud

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"sync"
	"testing"
	"time"
)

func loopWorld() *World {
	return NewWorld("torg",
		&Room{ID: "torg", Name: "Square", Description: "A square.", Terrain: "town", Exits: map[string]string{"north": "krog"}},
		&Room{ID: "krog", Name: "Tavern", Description: "A tavern.", Terrain: "inside", Y: -1, Exits: map[string]string{"south": "torg"}},
	)
}

func TestGameTime(t *testing.T) {
	tests := []struct {
		ticks             int64
		day, hour, minute int
	}{
		{0, 1, 8, 0},
		{1, 1, 8, 1},
		{59, 1, 8, 59},
		{60, 1, 9, 0},
		{16*60 - 1, 1, 23, 59},
		{16 * 60, 2, 0, 0},
		{16*60 + 24*60, 3, 0, 0},
	}
	for _, tt := range tests {
		d, h, m := GameTime(tt.ticks)
		if d != tt.day || h != tt.hour || m != tt.minute {
			t.Errorf("GameTime(%d) = %d, %d, %d; %s %d, %d, %d", tt.ticks, d, h, m, L("väntade", "expected", "期待値"), tt.day, tt.hour, tt.minute)
		}
	}
}

func TestTicksAdvance(t *testing.T) {
	srv := NewServerWithOptions(Options{World: loopWorld(), Tick: 2 * time.Millisecond})
	startServer(t, srv)
	time.Sleep(150 * time.Millisecond)
	if n := srv.Ticks(); n < 10 {
		t.Errorf(L("efter 150 ms med 2 ms tick har världen bara tickat %d gånger – körs spelloopen?",
			"after 150 ms with a 2 ms tick the world has only ticked %d times – is the game loop running?",
			"2 ms ティックで 150 ms 経っても、世界は %d 回しか進んでいません。ゲームループは動いていますか？"), n)
	}
}

func TestTimeCommand(t *testing.T) {
	// A tick of one hour: the clock does not move during the test.
	srv := NewServerWithOptions(Options{World: loopWorld(), Tick: time.Hour})
	addr := startServer(t, srv)
	c := login(t, addr, "Anna")
	c.drain()
	c.send("time")
	c.expect(fmt.Sprintf(MsgTime, 1, 8, 0))
}

func TestDefaultOptions(t *testing.T) {
	srv := NewServerWithOptions(Options{World: loopWorld()})
	addr := startServer(t, srv)
	c := login(t, addr, "Anna")
	c.expectContains("Square")
}

func TestShutdown(t *testing.T) {
	srv := NewServerWithOptions(Options{World: loopWorld(), Tick: 5 * time.Millisecond})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(l) }()
	addr := l.Addr().String()

	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	waiting := connect(t, addr) // connected but not logged in yet
	waiting.expectContains(MsgAskName)
	anna.drain()
	bertil.drain()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf(L("Shutdown gav fel: %v", "Shutdown returned an error: %v", "Shutdown がエラーを返しました: %v"), err)
	}
	for _, c := range []*testClient{anna, bertil, waiting} {
		c.expectContains(MsgShutdown)
		c.expectClosed()
	}
	select {
	case err := <-served:
		if err != ErrServerClosed {
			t.Errorf(L("Serve ska returnera ErrServerClosed efter Shutdown, returnerade %v",
				"Serve should return ErrServerClosed after Shutdown, returned %v",
				"Shutdown の後、Serve は ErrServerClosed を返すべきですが、%v でした"), err)
		}
	case <-time.After(testTimeout):
		t.Error(L("Serve returnerade inte efter Shutdown", "Serve did not return after Shutdown", "Shutdown の後も Serve が戻りませんでした"))
	}
	if _, err := net.DialTimeout("tcp", addr, 200*time.Millisecond); err == nil {
		t.Error(L("servern tar fortfarande emot anslutningar efter Shutdown", "the server still accepts connections after Shutdown", "Shutdown の後もサーバーが接続を受け付けています"))
	}
}

func TestShutdownWithoutServe(t *testing.T) {
	srv := NewServerWithOptions(Options{World: loopWorld()})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Errorf(L("Shutdown på en server som aldrig startat ska fungera, gav %v",
			"Shutdown on a server that was never started should work, returned %v",
			"一度も起動していないサーバーの Shutdown も成功するべきですが、%v でした"), err)
	}
}

func TestManyPlayersAtOnce(t *testing.T) {
	srv := NewServerWithOptions(Options{World: loopWorld(), Tick: time.Millisecond})
	addr := startServer(t, srv)
	const n = 10
	var wg sync.WaitGroup
	errs := make(chan string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "Player" + string(rune('a'+i))
			conn, err := net.DialTimeout("tcp", addr, testTimeout)
			if err != nil {
				errs <- err.Error()
				return
			}
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(10 * time.Second))
			fmt.Fprintf(conn, "%s\r\n", name)
			for j := 0; j < 30; j++ {
				for _, cmd := range []string{"say hi", "n", "s", "look", "who", "time"} {
					if _, err := fmt.Fprintf(conn, "%s\r\n", cmd); err != nil {
						errs <- err.Error()
						return
					}
				}
			}
			fmt.Fprint(conn, "quit\r\n")
			// Read until the server closes the connection after quit.
			buf := make([]byte, 64*1024)
			for {
				if _, err := conn.Read(buf); err != nil {
					break
				}
			}
		}(i)
	}
	finished := make(chan struct{})
	go func() { wg.Wait(); close(finished) }()
	select {
	case <-finished:
	case <-time.After(12 * time.Second):
		t.Fatal(L("10 spelare som skickar kommandon samtidigt blev inte klara – låser sig spelloopen?",
			"10 players sending commands at the same time did not finish – does the game loop get stuck?",
			"10 人が同時にコマンドを送ったら終わりませんでした。ゲームループが止まっていませんか？"))
	}
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	// The server still answers.
	c := login(t, addr, "Last")
	c.drain()
	c.send("time")
	if line, _ := c.readLine(testTimeout); !clockRe.MatchString(line) {
		t.Errorf(L("time ska visa klockan, fick %q", "time should show the clock, got %q", "time は時刻を表示するべきですが、%q でした"), line)
	}
}

var clockRe = regexp.MustCompile(`\d\d:\d\d`)
