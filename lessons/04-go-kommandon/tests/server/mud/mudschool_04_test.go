package mud

import (
	"strings"
	"testing"
	"time"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		line string
		verb string
		args []string
		rest string
	}{
		{"look", "look", nil, ""},
		{"LOOK", "look", nil, ""},
		{"  look  ", "look", nil, ""},
		{"", "", nil, ""},
		{"     ", "", nil, ""},
		{"get sword", "get", []string{"sword"}, "sword"},
		{"Get Sword from  chest", "get", []string{"Sword", "from", "chest"}, "Sword from  chest"},
		{"say Hej på dig!", "say", []string{"Hej", "på", "dig!"}, "Hej på dig!"},
		{"'hej allihop", "say", []string{"hej", "allihop"}, "hej allihop"},
		{"n", "north", nil, ""},
		{"s", "south", nil, ""},
		{"e", "east", nil, ""},
		{"w", "west", nil, ""},
		{"u", "up", nil, ""},
		{"d", "down", nil, ""},
		{"l", "look", nil, ""},
		{"L", "look", nil, ""},
		{"?", "help", nil, ""},
		{"q", "quit", nil, ""},
	}
	for _, tt := range tests {
		got := ParseCommand(tt.line)
		if got.Verb != tt.verb || strings.Join(got.Args, "|") != strings.Join(tt.args, "|") || got.Rest != tt.rest {
			t.Errorf("ParseCommand(%q)\n  fick    Verb=%q Args=%q Rest=%q\n  väntade Verb=%q Args=%q Rest=%q",
				tt.line, got.Verb, got.Args, got.Rest, tt.verb, tt.args, tt.rest)
		}
	}
}

func TestHelpListsCommands(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("help")
	line := c.expectContains("Kommandon")
	for _, verb := range []string{"echo", "help", "quit"} {
		if !strings.Contains(line, verb) {
			t.Errorf("help ska nämna %q, men skrev %q", verb, line)
		}
	}
}

func TestRegisteredCommandsAppearInHelp(t *testing.T) {
	srv := NewServer()
	srv.Register("dance", func(s *Server, p *Player, cmd Command) { p.Send("Du dansar!") })
	addr := startServer(t, srv)
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("dance")
	c.expect("Du dansar!")
	c.send("?")
	if line := c.expectContains("Kommandon"); !strings.Contains(line, "dance") {
		t.Errorf("help ska lista alla registrerade kommandon, men skrev %q", line)
	}
}

func TestEchoAndUnknown(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("echo Hej  där")
	c.expect("Hej  där")
	c.send("FLYG")
	c.expect(`Okänt kommando "flyg". Skriv help för hjälp.`)
}

func TestEmptyLineIsIgnored(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("")
	c.send("   ")
	c.expectSilence(200 * time.Millisecond)
}

func TestQuit(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("quit")
	c.expect("Hej då!")
	c.expectClosed()
}
