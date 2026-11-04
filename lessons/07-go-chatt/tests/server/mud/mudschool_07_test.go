package mud

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSay(t *testing.T) {
	addr := startServer(t, NewServer())
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.expectContains(fmt.Sprintf(MsgJoined, "Bertil"))

	anna.send("say Hello Bertil!")
	anna.expect(fmt.Sprintf(MsgYouSay, "Hello Bertil!"))
	bertil.expect(fmt.Sprintf(MsgSays, "Anna", "Hello Bertil!"))

	bertil.send("'hi hi")
	bertil.expect(fmt.Sprintf(MsgYouSay, "hi hi"))
	anna.expect(fmt.Sprintf(MsgSays, "Bertil", "hi hi"))

	anna.send("say")
	anna.expect(MsgSayWhat)
}

func TestJoinAndLeaveAreAnnounced(t *testing.T) {
	addr := startServer(t, NewServer())
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.expect(fmt.Sprintf(MsgJoined, "Bertil"))

	bertil.send("quit")
	anna.expect(fmt.Sprintf(MsgLeft, "Bertil"))

	cecilia := login(t, addr, "Cecilia")
	anna.expect(fmt.Sprintf(MsgJoined, "Cecilia"))
	cecilia.conn.Close() // disconnect without quit
	anna.expect(fmt.Sprintf(MsgLeft, "Cecilia"))
}

func TestQuitStillSaysGoodbye(t *testing.T) {
	addr := startServer(t, NewServer())
	c := login(t, addr, "Anna")
	c.send("quit")
	c.expect(MsgGoodbye)
	c.expectClosed()
}

func TestSlowClientDoesNotBlockOthers(t *testing.T) {
	addr := startServer(t, NewServer())
	// Lazy Ludvig logs in but never reads anything again.
	login(t, addr, "Ludvig")
	chatty := login(t, addr, "Pratsam")

	youSay := strings.SplitN(MsgYouSay, "%s", 2)[0]
	long := strings.Repeat("bla", 300)
	start := time.Now()
	// ~8 MB of text to Ludvig – far more than fits in the network buffers.
	for i := 0; i < 9000; i++ {
		chatty.send("say " + long)
		got, ok := chatty.readLine(testTimeout)
		if !ok {
			t.Fatal(L("servern stängde Pratsams anslutning", "the server closed Pratsam's connection", "サーバーが Pratsam の接続を閉じました"))
		}
		if !strings.HasPrefix(got, youSay) {
			continue
		}
		if time.Since(start) > 20*time.Second {
			t.Fatal(L("det går för långsamt – blockerar servern på Ludvig?",
				"this is far too slow – is the server blocking on Ludvig?",
				"遅すぎます。サーバーが Ludvig で詰まっていませんか？"))
		}
	}
	chatty.seen = nil
}
