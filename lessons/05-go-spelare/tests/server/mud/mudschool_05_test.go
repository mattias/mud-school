package mud

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name string
		want error
	}{
		{"Anna", nil},
		{"bo", nil},
		{"Åsa", nil},
		{"Sixtenkarlssonxx", nil}, // 16 tecken
		{"a", ErrNameTooShort},
		{"Ö", ErrNameTooShort}, // en bokstav men två bytes!
		{"", ErrNameTooShort},
		{"Sixtenkarlssonxxx", ErrNameTooLong},
		{"Anna1", ErrNameInvalid},
		{"Anna Berg", ErrNameInvalid},
		{"x_x", ErrNameInvalid},
	}
	for _, tt := range tests {
		err := ValidateName(tt.name)
		if tt.want == nil && err != nil {
			t.Errorf("ValidateName(%q) = %v, väntade nil", tt.name, err)
		}
		if tt.want != nil && !errors.Is(err, tt.want) {
			t.Errorf("ValidateName(%q) = %v, väntade %v", tt.name, err, tt.want)
		}
	}
}

func TestNormalizeName(t *testing.T) {
	for in, want := range map[string]string{
		"anna": "Anna", "ANNA": "Anna", "aNNa": "Anna", "åsa": "Åsa", "ÖRJAN": "Örjan",
	} {
		if got := NormalizeName(in); got != want {
			t.Errorf("NormalizeName(%q) = %q, väntade %q", in, got, want)
		}
	}
}

func TestLoginAsksAgainOnInvalidName(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.expect("Vad heter du?")
	c.send("x")
	c.expectContains("minst 2 tecken")
	c.expect("Vad heter du?")
	c.send("r2d2")
	c.expectContains("bokstäver")
	c.expect("Vad heter du?")
	c.send("anna")
	c.expect("Välkommen, Anna!")
}

func TestNameMustBeUnique(t *testing.T) {
	addr := startServer(t, NewServer())
	login(t, addr, "Anna")

	c := connect(t, addr)
	c.expectContains("Vad heter du?")
	c.send("ANNA")
	c.expectContains("upptaget")
	c.expect("Vad heter du?")
	c.send("Bertil")
	c.expect("Välkommen, Bertil!")
}

func TestWho(t *testing.T) {
	addr := startServer(t, NewServer())
	bertil := login(t, addr, "Bertil")
	anna := login(t, addr, "Anna")
	anna.send("who")
	line := anna.expectContains("online")
	if !strings.Contains(line, "(2)") || !strings.Contains(line, "Anna, Bertil") {
		t.Errorf("who ska skriva antal och sorterade namn, t.ex. %q, men skrev %q",
			"Spelare online (2): Anna, Bertil", line)
	}

	bertil.send("quit")
	bertil.expectClosed()
	anna.send("who")
	if line := anna.expectContains("online"); !strings.Contains(line, "(1)") || strings.Contains(line, "Bertil") {
		t.Errorf("Bertil har loggat ut men who skrev %q", line)
	}
}

func TestNameIsFreeAfterDisconnect(t *testing.T) {
	addr := startServer(t, NewServer())
	first := login(t, addr, "Cecilia")
	first.conn.Close()

	// Servern behöver ett ögonblick för att märka att anslutningen stängts.
	for i := 0; ; i++ {
		c := connect(t, addr)
		c.expectContains("Vad heter du?")
		c.send("Cecilia")
		line, _ := c.readLine(testTimeout)
		if strings.Contains(line, "Välkommen, Cecilia!") {
			return
		}
		if i == 20 {
			t.Fatalf("namnet Cecilia blev aldrig ledigt efter att spelaren kopplat ner (sista svar: %q)", line)
		}
		c.conn.Close()
	}
}
