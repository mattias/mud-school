package mud

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// A world with a closed-off tutorial area (hall → yard) below the square.
const tutorialWorldJSON = `{
  "start": "square",
  "rooms": [
    {"id": "square", "name": "Square", "description": "A square.", "terrain": "town"},
    {"id": "hall", "name": "Hall", "description": "A hall.", "terrain": "inside", "z": -1,
     "exits": {"north": "yard"}},
    {"id": "yard", "name": "Yard", "description": "A yard.", "terrain": "field", "y": -1, "z": -1,
     "exits": {"south": "hall"}}
  ],
  "items": [
    {"id": "torch", "name": "a torch", "description": "A torch.", "keywords": ["torch"], "room": "yard"}
  ],
  "tutorial": {
    "start": "hall",
    "steps": [
      {"text": "Type look.", "command": "look"},
      {"text": "Go north.", "room": "yard"},
      {"text": "Get the torch.", "item": "torch"},
      {"text": "Say something.", "command": "say"}
    ],
    "done": "Well done!"
  }
}`

func step37(n int, text string) string { return fmt.Sprintf(MsgTutorialStep, n, 4, text) }

func newTutorialServer(t *testing.T, st *Store) (*Server, string) {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(tutorialWorldJSON))
	if err != nil {
		t.Fatalf(L("världen med en introduktion laddas inte: %v", "the world with a tutorial doesn't load: %v", "チュートリアル付きの世界が読み込めません: %v"), err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, Store: st})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return srv, addr
}

func TestNewPlayersStartTheTutorial(t *testing.T) {
	_, addr := newTutorialServer(t, nil)
	c := login(t, addr, "Anna")
	c.expectContains("Hall")
	c.expectContains(step37(1, "Type look."))
}

func TestTheTutorialStepByStep(t *testing.T) {
	_, addr := newTutorialServer(t, nil)
	c := login(t, addr, "Anna")
	c.expectContains(step37(1, "Type look."))
	c.send("look")
	c.expectContains(step37(2, "Go north."))
	c.send("go north")
	c.expectContains("Yard")
	c.expectContains(step37(3, "Get the torch."))
	c.send("get torch")
	c.expectContains(step37(4, "Say something."))
	c.send("say hello")
	c.expectContains("Well done!")
	c.expectContains("Square") // out in the world
	c.send("tutorial")
	c.expectContains(MsgTutorialNotIn)
	c.drain()
	c.send("inventory")
	if line, _ := c.readLine(testTimeout); strings.Contains(line, "torch") {
		t.Errorf(L("facklan från introduktionen ska lämnas tillbaka när man går ut: %q",
			"the torch from the tutorial should be given back when you leave it: %q",
			"チュートリアルのたいまつは、出るときに返すべきです: %q"), line)
	}

	// The next new player finds the torch where it was.
	b := login(t, addr, "Bertil")
	b.drain()
	b.send("go north")
	b.expectContains(fmt.Sprintf(MsgItemsHere, "a torch"))
}

func TestCommandsThatDontMatchDontAdvance(t *testing.T) {
	_, addr := newTutorialServer(t, nil)
	c := login(t, addr, "Anna")
	c.expectContains(step37(1, "Type look."))
	c.send("who")
	c.send("dance") // unknown – not a command that has run
	c.send("tutorial")
	c.expectContains(step37(1, "Type look."))
}

func TestSkipTheTutorial(t *testing.T) {
	_, addr := newTutorialServer(t, nil)
	c := login(t, addr, "Anna")
	c.expectContains(step37(1, "Type look."))
	c.send("tutorial skip")
	c.expectContains(MsgTutorialSkipped)
	c.expectContains("Square")
	c.send("tutorial")
	c.expectContains(MsgTutorialNotIn)
}

func TestNoTutorialNoProblem(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(`{"start": "square", "rooms": [
		{"id": "square", "name": "Square", "description": "A square.", "terrain": "town"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	c := login(t, addr, "Anna")
	c.expectContains("Square")
	c.send("tutorial")
	c.expectContains(MsgTutorialNotIn)
}

func init() { PasswordCost = bcrypt.MinCost }

func loginPw37(t *testing.T, addr, name string, first bool) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expectContains(MsgAskName)
	c.send(name)
	if first {
		c.expectContains(fmt.Sprintf(MsgChoosePassword, name))
	} else {
		c.expectContains(MsgAskPassword)
	}
	c.send("secret")
	c.expectContains(fmt.Sprintf(MsgWelcomeName, name))
	return c
}

func TestTutorialProgressIsSaved(t *testing.T) {
	st, err := OpenStore(filepath.Join(t.TempDir(), "mud.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	_, addr := newTutorialServer(t, st)

	c := loginPw37(t, addr, "Anna", true)
	c.expectContains(step37(1, "Type look."))
	c.send("look")
	c.expectContains(step37(2, "Go north."))
	c.send("quit")
	c.expectClosed()

	sp, found, err := st.LoadPlayer("Anna")
	if err != nil || !found || sp.Tutorial != 2 {
		t.Fatalf(L("efter att ha loggat ut på steg 2 ska Tutorial vara 2 i databasen, fick %+v (hittad %v, fel %v)",
			"after logging out on step 2, Tutorial should be 2 in the database, got %+v (found %v, error %v)",
			"ステップ 2 でログアウトしたら、データベースの Tutorial は 2 のはずですが %+v（found %v、エラー %v）"), sp, found, err)
	}

	c = loginPw37(t, addr, "Anna", false)
	c.expectContains("Hall")
	c.expectContains(step37(2, "Go north."))
	c.send("tutorial skip")
	c.expectContains("Square")
	c.send("quit")
	c.expectClosed()

	c = loginPw37(t, addr, "Anna", false)
	c.expectContains("Square")
	c.send("tutorial")
	c.expectContains(MsgTutorialNotIn)
}

func TestOldPlayersDontGetTheTutorial(t *testing.T) {
	st, err := OpenStore(filepath.Join(t.TempDir(), "mud.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	// Bertil played before the tutorial existed: his row has the default 0.
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err := st.CreateAccount("Bertil", hash); err != nil {
		t.Fatal(err)
	}
	if _, err := st.db.Exec(`INSERT INTO players (name, room, hp) VALUES ('Bertil', 'square', 10)`); err != nil {
		t.Fatalf(L("players ska fortfarande gå att fylla utan kolumnen tutorial (DEFAULT 0): %v",
			"players should still accept a row without the tutorial column (DEFAULT 0): %v",
			"tutorial 列なしでも players に行を入れられるべきです（DEFAULT 0）: %v"), err)
	}
	_, addr := newTutorialServer(t, st)
	c := loginPw37(t, addr, "Bertil", false)
	c.expectContains("Square")
	c.send("tutorial")
	c.expectContains(MsgTutorialNotIn)
}

func TestTheCommandEvent(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(tutorialWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	verbs := make(chan string, 10)
	srv.On(EventCommand, func(e *Event) {
		if e.Player == nil || e.Room == nil {
			t.Errorf(L("EventCommand ska ha Player och Room", "EventCommand should have Player and Room", "EventCommand には Player と Room が必要です"))
		}
		verbs <- e.Text
	})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	c := login(t, addr, "Anna")
	c.drain()
	c.send("l")
	c.send("dance")
	c.send("who")
	c.drain()
	got := []string{}
	for len(verbs) > 0 {
		got = append(got, <-verbs)
	}
	if strings.Join(got, ",") != "look,who" {
		t.Errorf(L("EventCommand ska komma efter varje kommando som fanns, med verbet (alias expanderade) i Text: fick %q, väntade [look who]",
			"EventCommand should come after every command that exists, with the verb (aliases expanded) in Text: got %q, expected [look who]",
			"EventCommand は存在するコマンドのたびに、動詞（別名は展開済み）を Text に入れて来るべきです: %q（期待 [look who]）"), got)
	}
}

func TestTutorialValidation(t *testing.T) {
	bad := `{
  "start": "square",
  "rooms": [
    {"id": "square", "name": "Square", "description": "A square.", "terrain": "town"},
    {"id": "hall", "name": "Hall", "description": "A hall.", "terrain": "inside", "z": -1,
     "exits": {"up": "square"}},
    {"id": "shop", "name": "Shop", "description": "A shop.", "terrain": "inside", "x": 5}
  ],
  "items": [
    {"id": "coin", "name": "a coin", "description": "A coin.", "keywords": ["coin"], "room": "shop"}
  ],
  "tutorial": {
    "start": "hall",
    "steps": [
      {"text": "Two at once.", "command": "look", "room": "hall"},
      {"text": "Nowhere.", "room": "attic"},
      {"text": "Not here.", "item": "coin"},
      {"command": "look"}
    ],
    "done": "Done."
  }
}`
	_, err := LoadWorld(strings.NewReader(bad))
	if err == nil {
		t.Fatal(L("en introduktion med fel ska ge ett fel", "a tutorial with mistakes should give an error", "誤りのあるチュートリアルはエラーになるべきです"))
	}
	for _, want := range []string{
		fmt.Sprintf(MsgErrTutorialOpen, "square"),
		fmt.Sprintf(MsgErrTutorialStep, 1),
		fmt.Sprintf(MsgErrTutorialRoom, 2, "attic"),
		fmt.Sprintf(MsgErrTutorialItem, 3, "coin"),
		fmt.Sprintf(MsgErrTutorialStep, 4),
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf(L("felet ska innehålla %q (alla fel på en gång):\n%v", "the error should contain %q (all the mistakes at once):\n%v", "エラーには %q が含まれるべきです（すべての誤りを一度に）:\n%v"), want, err)
		}
	}
	_, err = LoadWorld(strings.NewReader(`{"start": "square", "rooms": [
		{"id": "square", "name": "Square", "description": "A square.", "terrain": "town"}],
		"tutorial": {"start": "cellar", "steps": [{"text": "Hi.", "command": "look"}], "done": "Done."}}`))
	if err == nil || !strings.Contains(err.Error(), fmt.Sprintf(MsgErrTutorialStart, "cellar")) {
		t.Errorf(L("ett startrum som saknas ska ge %q, fick %v", "a missing start room should give %q, got %v", "存在しない開始の部屋は %q になるべきですが %v"), fmt.Sprintf(MsgErrTutorialStart, "cellar"), err)
	}
}

func TestTheDefaultWorldHasATutorial(t *testing.T) {
	w := DefaultWorld()
	if w.Tutorial == nil || len(w.Tutorial.Steps) == 0 {
		t.Fatal(L("standardvärlden ska ha en introduktion (tutorial.json)", "the default world should have a tutorial (tutorial.json)", "標準の世界にはチュートリアル（tutorial.json）が必要です"))
	}
	if w.Room(w.Tutorial.Start) == nil {
		t.Errorf(L("introduktionens startrum %q finns inte", "the tutorial's start room %q doesn't exist", "チュートリアルの開始の部屋 %q がありません"), w.Tutorial.Start)
	}
}

func TestTheTutorialSurvivesBuilding(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(tutorialWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	var f worldFile
	srv.call(func() { f = srv.worldSnapshot() })
	shutdown(t, srv)
	if f.Tutorial == nil || f.Tutorial.Start != "hall" {
		t.Errorf(L("worldSnapshot (lektion 36) ska ta med introduktionen, annars försvinner den när någon bygger",
			"worldSnapshot (lesson 36) should include the tutorial, or it disappears when someone builds",
			"worldSnapshot（レッスン 36）はチュートリアルも含めるべきです。でないと誰かが建築したときに消えます"))
	}
}
