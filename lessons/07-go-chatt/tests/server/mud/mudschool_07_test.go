package mud

import (
	"strings"
	"testing"
	"time"
)

func TestSay(t *testing.T) {
	addr := startServer(t, NewServer())
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.expectContains("Bertil har anslutit.")

	anna.send("say Hej Bertil!")
	anna.expect("Du säger: Hej Bertil!")
	bertil.expect("Anna säger: Hej Bertil!")

	bertil.send("'hej hej")
	bertil.expect("Du säger: hej hej")
	anna.expect("Bertil säger: hej hej")

	anna.send("say")
	anna.expect("Säg vad?")
}

func TestJoinAndLeaveAreAnnounced(t *testing.T) {
	addr := startServer(t, NewServer())
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.expect("Bertil har anslutit.")

	bertil.send("quit")
	anna.expect("Bertil har lämnat.")

	cecilia := login(t, addr, "Cecilia")
	anna.expect("Cecilia har anslutit.")
	cecilia.conn.Close() // koppla ner utan quit
	anna.expect("Cecilia har lämnat.")
}

func TestQuitStillSaysGoodbye(t *testing.T) {
	addr := startServer(t, NewServer())
	c := login(t, addr, "Anna")
	c.send("quit")
	c.expect("Hej då!")
	c.expectClosed()
}

func TestSlowClientDoesNotBlockOthers(t *testing.T) {
	addr := startServer(t, NewServer())
	// Lata Ludvig loggar in men läser aldrig något mer.
	login(t, addr, "Ludvig")
	pratsam := login(t, addr, "Pratsam")

	long := strings.Repeat("bla", 300)
	start := time.Now()
	// ~8 MB text till Ludvig – mycket mer än vad som ryms i nätverksbuffertarna.
	for i := 0; i < 9000; i++ {
		pratsam.send("say " + long)
		got, ok := pratsam.readLine(testTimeout)
		if !ok {
			t.Fatal("servern stängde Pratsams anslutning")
		}
		if !strings.HasPrefix(got, "Du säger:") {
			continue
		}
		if time.Since(start) > 20*time.Second {
			t.Fatal("det går för långsamt – blockerar servern på Ludvig?")
		}
	}
	pratsam.seen = nil
}
