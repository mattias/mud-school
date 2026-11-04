package mud

import (
	"errors"
	"fmt"
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
		{"さくら", nil},
		{"Sixtenkarlssonxx", nil}, // 16 characters
		{"a", ErrNameTooShort},
		{"Ö", ErrNameTooShort}, // one letter but two bytes!
		{"空", ErrNameTooShort}, // one letter but three bytes!
		{"", ErrNameTooShort},
		{"Sixtenkarlssonxxx", ErrNameTooLong},
		{"Anna1", ErrNameInvalid},
		{"Anna Berg", ErrNameInvalid},
		{"x_x", ErrNameInvalid},
	}
	for _, tt := range tests {
		err := ValidateName(tt.name)
		if tt.want == nil && err != nil {
			t.Errorf(L("ValidateName(%q) = %v, väntade nil", "ValidateName(%q) = %v, expected nil", "ValidateName(%q) = %v ですが、nil を期待しました"), tt.name, err)
		}
		if tt.want != nil && !errors.Is(err, tt.want) {
			t.Errorf(L("ValidateName(%q) = %v, väntade %v", "ValidateName(%q) = %v, expected %v", "ValidateName(%q) = %v ですが、%v を期待しました"), tt.name, err, tt.want)
		}
	}
}

func TestNormalizeName(t *testing.T) {
	for in, want := range map[string]string{
		"anna": "Anna", "ANNA": "Anna", "aNNa": "Anna", "åsa": "Åsa", "ÖRJAN": "Örjan", "さくら": "さくら",
	} {
		if got := NormalizeName(in); got != want {
			t.Errorf(L("NormalizeName(%q) = %q, väntade %q", "NormalizeName(%q) = %q, expected %q", "NormalizeName(%q) = %q ですが、%q を期待しました"), in, got, want)
		}
	}
}

func TestLoginAsksAgainOnInvalidName(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.expect(MsgAskName)
	c.send("x")
	c.expectContains(MsgErrNameTooShort)
	c.expect(MsgAskName)
	c.send("r2d2")
	c.expectContains(MsgErrNameInvalid)
	c.expect(MsgAskName)
	c.send("anna")
	c.expect(fmt.Sprintf(MsgWelcomeName, "Anna"))
}

func TestNameMustBeUnique(t *testing.T) {
	addr := startServer(t, NewServer())
	login(t, addr, "Anna")

	c := connect(t, addr)
	c.expectContains(MsgAskName)
	c.send("ANNA")
	c.expectContains(MsgErrNameTaken)
	c.expect(MsgAskName)
	c.send("Bertil")
	c.expect(fmt.Sprintf(MsgWelcomeName, "Bertil"))
}

func TestWho(t *testing.T) {
	addr := startServer(t, NewServer())
	bertil := login(t, addr, "Bertil")
	anna := login(t, addr, "Anna")
	anna.send("who")
	want := fmt.Sprintf(MsgWho, 2, "Anna"+MsgListSep+"Bertil")
	anna.expect(want)

	bertil.send("quit")
	bertil.expectClosed()
	anna.send("who")
	anna.expect(fmt.Sprintf(MsgWho, 1, "Anna"))
}

func TestNameIsFreeAfterDisconnect(t *testing.T) {
	addr := startServer(t, NewServer())
	first := login(t, addr, "Cecilia")
	first.conn.Close()

	// The server needs a moment to notice that the connection was closed.
	welcome := fmt.Sprintf(MsgWelcomeName, "Cecilia")
	for i := 0; ; i++ {
		c := connect(t, addr)
		c.expectContains(MsgAskName)
		c.send("Cecilia")
		line, _ := c.readLine(testTimeout)
		if strings.Contains(line, welcome) {
			return
		}
		if i == 20 {
			t.Fatalf(L("namnet Cecilia blev aldrig ledigt efter att spelaren kopplat ner (sista svar: %q)",
				"the name Cecilia never became free after the player disconnected (last reply: %q)",
				"プレイヤーが切断した後も名前 Cecilia が空きませんでした（最後の応答: %q）"), line)
		}
		c.conn.Close()
	}
}
