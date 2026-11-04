package mud

import (
	"fmt"
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
		{"say Hello there!", "say", []string{"Hello", "there!"}, "Hello there!"},
		{"say こんにちは 世界", "say", []string{"こんにちは", "世界"}, "こんにちは 世界"},
		{"'hello all", "say", []string{"hello", "all"}, "hello all"},
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
			t.Errorf(L("ParseCommand(%q)\n  fick    Verb=%q Args=%q Rest=%q\n  väntade Verb=%q Args=%q Rest=%q",
				"ParseCommand(%q)\n  got      Verb=%q Args=%q Rest=%q\n  expected Verb=%q Args=%q Rest=%q",
				"ParseCommand(%q)\n  実際  Verb=%q Args=%q Rest=%q\n  期待  Verb=%q Args=%q Rest=%q"),
				tt.line, got.Verb, got.Args, got.Rest, tt.verb, tt.args, tt.rest)
		}
	}
}

// commandsPrefix is MsgCommands up to the list, e.g. "Commands: ".
var commandsPrefix = strings.SplitN(MsgCommands, "%s", 2)[0]

func TestHelpListsCommands(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("help")
	line := c.expectContains(commandsPrefix)
	for _, verb := range []string{"echo", "help", "quit"} {
		if !strings.Contains(line, verb) {
			t.Errorf(L("help ska nämna %q, men skrev %q", "help should mention %q, but wrote %q", "help は %q を含むべきですが、出力は %q でした"), verb, line)
		}
	}
	want := fmt.Sprintf(MsgCommands, strings.Join([]string{"echo", "help", "quit"}, MsgListSep))
	if line != want {
		t.Errorf(L("help ska skriva verben sorterade och åtskilda med MsgListSep\n  fick:    %q\n  väntade: %q",
			"help should write the verbs sorted and separated by MsgListSep\n  got:      %q\n  expected: %q",
			"help は動詞をソートし MsgListSep で区切って表示するべきです\n  実際: %q\n  期待: %q"), line, want)
	}
}

func TestRegisteredCommandsAppearInHelp(t *testing.T) {
	srv := NewServer()
	srv.Register("dance", func(s *Server, p *Player, cmd Command) { p.Send("*dance*") })
	addr := startServer(t, srv)
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("dance")
	c.expect("*dance*")
	c.send("?")
	if line := c.expectContains(commandsPrefix); !strings.Contains(line, "dance") {
		t.Errorf(L("help ska lista alla registrerade kommandon, men skrev %q",
			"help should list every registered command, but wrote %q",
			"help は登録済みのすべてのコマンドを表示するべきですが、出力は %q でした"), line)
	}
}

func TestEchoAndUnknown(t *testing.T) {
	addr := startServer(t, NewServer())
	c := connect(t, addr)
	c.expect(Welcome)
	c.send("echo Hello  there")
	c.expect("Hello  there")
	c.send("FLY")
	c.expect(fmt.Sprintf(MsgUnknownCommand, "fly"))
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
	c.expect(MsgGoodbye)
	c.expectClosed()
}
