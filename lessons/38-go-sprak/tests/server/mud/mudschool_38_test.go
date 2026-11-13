package mud

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// A world in English with Swedish translations of the hall and the tutorial.
const languageWorldJSON = `{
  "start": "square",
  "language": "en",
  "rooms": [
    {"id": "square", "name": "Square", "description": "A square.", "terrain": "town"},
    {"id": "hall", "name": "Hall", "description": "A hall.", "terrain": "inside", "z": -1,
     "translations": {"sv": {"name": "Salen", "description": "En sal."}}}
  ],
  "tutorial": {
    "start": "hall",
    "steps": [
      {"text": {"en": "Hello {{.Name}}, type look.", "sv": "Hej {{.Name}}, skriv look."}, "command": "look"},
      {"text": "Type who ({{.Step}}/{{.Steps}}).", "command": "who"}
    ],
    "done": {"en": "Done!", "sv": "Klart!"}
  }
}`

func step38(n int, text string) string { return fmt.Sprintf(MsgTutorialStep, n, 2, text) }

func newLanguageServer(t *testing.T, st *Store) string {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(languageWorldJSON))
	if err != nil {
		t.Fatalf(L("världen med översättningar laddas inte: %v", "the world with translations doesn't load: %v", "翻訳付きの世界が読み込めません: %v"), err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, Store: st})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return addr
}

func TestTextFromJSON(t *testing.T) {
	var v struct {
		A Text `json:"a"`
		B Text `json:"b"`
	}
	if err := json.Unmarshal([]byte(`{"a": "plain", "b": {"en": "Hi", "sv": "Hej"}}`), &v); err != nil {
		t.Fatalf(L("Text ska kunna läsas både som sträng och som objekt: %v", "Text should decode from both a string and an object: %v", "Text は文字列からもオブジェクトからも読めるべきです: %v"), err)
	}
	if v.A[""] != "plain" || v.B["sv"] != "Hej" || v.B["en"] != "Hi" {
		t.Errorf(L("fick %v och %v; väntade {\"\": plain} och {en: Hi, sv: Hej}", "got %v and %v; expected {\"\": plain} and {en: Hi, sv: Hej}", "%v と %v になりました。{\"\": plain} と {en: Hi, sv: Hej} を期待しました"), v.A, v.B)
	}
	if err := json.Unmarshal([]byte(`{"a": 42}`), &v); err == nil || !strings.Contains(err.Error(), MsgErrText) {
		t.Errorf(L("ett tal är ingen Text – väntade ett fel med %q, fick %v", "a number is not a Text – expected an error with %q, got %v", "数値は Text ではありません。%q を含むエラーを期待しましたが %v"), MsgErrText, err)
	}
	out, _ := json.Marshal(v)
	if !strings.Contains(string(out), `"a":"plain"`) || !strings.Contains(string(out), `"sv":"Hej"`) {
		t.Errorf(L("en Text med bara en vanlig sträng ska sparas som en sträng igen, fick %s", "a Text with only a plain string should be saved as a string again, got %s", "普通の文字列だけの Text は文字列として保存されるべきです: %s"), out)
	}
}

func TestTextFallback(t *testing.T) {
	tx := Text{"en": "Hi", "sv": "Hej", "": "Plain"}
	cases := []struct {
		lang, fallback, want string
	}{
		{"sv", "en", "Hej"},
		{"ja", "en", "Hi"},
		{"ja", "fr", "Plain"},
		{"", "", "Plain"},
	}
	for _, c := range cases {
		if got := tx.In(c.lang, c.fallback); got != c.want {
			t.Errorf("Text%v.In(%q, %q) = %q, %s %q", tx, c.lang, c.fallback, got, L("väntade", "expected", "期待"), c.want)
		}
	}
	if got := (Text{"sv": "Hej", "en": "Hi"}).In("ja", "fr"); got != "Hi" {
		t.Errorf(L("utan träff ska In välja det första språket i bokstavsordning (en), fick %q", "with no match In should pick the first language in alphabetical order (en), got %q", "一致がなければ In はアルファベット順で最初の言語（en）を選ぶべきですが %q"), got)
	}
}

func TestTextTemplates(t *testing.T) {
	got := Text{"": "Hi {{.Name}}, step {{.Step}} of {{.Steps}}."}.Render("en", "en", TextData{Name: "Anna", Step: 2, Steps: 5})
	if got != "Hi Anna, step 2 of 5." {
		t.Errorf(L("Render ska köra texten som en text/template: fick %q", "Render should run the text as a text/template: got %q", "Render はテキストを text/template として実行するべきです: %q"), got)
	}
	bad := strings.Replace(languageWorldJSON, `"Type who ({{.Step}}/{{.Steps}})."`, `"Type {{.Nmae}}."`, 1)
	_, err := LoadWorld(strings.NewReader(bad))
	prefix, _, _ := strings.Cut(fmt.Sprintf(MsgErrTutorialText, 2, "\x00"), "\x00")
	if err == nil || !strings.Contains(err.Error(), prefix) {
		t.Errorf(L("en mall med ett fält som inte finns ({{.Nmae}}) ska ge ett fel när världen laddas, fick %v",
			"a template with a field that doesn't exist ({{.Nmae}}) should give an error when the world loads, got %v",
			"存在しないフィールド（{{.Nmae}}）のテンプレートは、世界を読み込むときにエラーになるべきです: %v"), err)
	}
}

func TestChooseALanguage(t *testing.T) {
	addr := newLanguageServer(t, nil)
	c := login(t, addr, "Anna")
	c.expectContains("Hall")
	c.expectContains(step38(1, "Hello Anna, type look."))
	c.send("language")
	c.expectContains(fmt.Sprintf(MsgLanguage, "en", "en"+MsgListSep+"sv"))
	c.send("language sv")
	c.expectContains(fmt.Sprintf(MsgLanguageSet, "sv"))
	c.expectContains("Salen")
	c.expectContains("En sal.")
	c.expectContains(step38(1, "Hej Anna, skriv look."))
	c.send("language fr")
	c.expectContains(fmt.Sprintf(MsgLanguageUnknown, "fr", "en"+MsgListSep+"sv"))
	c.send("look")
	c.expectContains(step38(2, "Type who (2/2).")) // only in one language
	c.send("who")
	c.expectContains("Klart!")
}

func TestOthersKeepTheirLanguage(t *testing.T) {
	addr := newLanguageServer(t, nil)
	a := login(t, addr, "Anna")
	a.drain()
	a.send("language sv")
	a.expectContains("Salen")
	b := login(t, addr, "Bertil")
	b.expectContains("Hall")
	b.expectContains(step38(1, "Hello Bertil, type look."))
}

func TestTheLanguageIsSaved(t *testing.T) {
	st, err := OpenStore(filepath.Join(t.TempDir(), "mud.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	addr := newLanguageServer(t, st)
	c := loginPw38(t, addr, "Anna", true)
	c.drain()
	c.send("language sv")
	c.expectContains("Salen")
	c.send("quit")
	c.expectClosed()
	if sp, _, _ := st.LoadPlayer("Anna"); sp.Language != "sv" {
		t.Errorf(L("språket ska sparas i databasen, Language = %q", "the language should be saved in the database, Language = %q", "言語はデータベースに保存されるべきです。Language = %q"), sp.Language)
	}
	c = loginPw38(t, addr, "Anna", false)
	c.expectContains("Salen")
}

func TestTheDefaultTutorialSpeaksThreeLanguages(t *testing.T) {
	w := DefaultWorld()
	langs := w.Languages()
	for _, want := range []string{"en", "ja", "sv"} {
		if !slices.Contains(langs, want) {
			t.Errorf(L("standardvärldens introduktion ska finnas på sv, en och ja (tutorial.all.json) – Languages() = %v",
				"the default world's tutorial should exist in sv, en and ja (tutorial.all.json) – Languages() = %v",
				"標準の世界のチュートリアルは sv・en・ja で用意されるべきです（tutorial.all.json）。Languages() = %v"), langs)
			break
		}
	}
	if w.Language != CourseLanguage {
		t.Errorf(L("standardvärldens Language ska vara CourseLanguage (%q), är %q", "the default world's Language should be CourseLanguage (%q), is %q", "標準の世界の Language は CourseLanguage（%q）のはずですが %q"), CourseLanguage, w.Language)
	}
}

func TestTheLanguageSurvivesBuilding(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(languageWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	var data []byte
	srv.call(func() { data, err = json.Marshal(srv.worldSnapshot()) })
	shutdown(t, srv)
	again, err := LoadWorld(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf(L("världen som worldSnapshot sparar ska gå att ladda igen: %v", "the world worldSnapshot saves should load again: %v", "worldSnapshot が保存した世界は再び読み込めるべきです: %v"), err)
	}
	if again.Language != "en" || again.Room("hall").Translations["sv"].Name != "Salen" || again.Tutorial.Done["sv"] != "Klart!" {
		t.Errorf(L("språket, översättningarna och introduktionens texter ska följa med när världen sparas",
			"the language, the translations and the tutorial's texts should be kept when the world is saved",
			"世界を保存するとき、言語・翻訳・チュートリアルのテキストも残すべきです"))
	}
}

func init() { PasswordCost = bcrypt.MinCost }

func loginPw38(t *testing.T, addr, name string, first bool) *testClient {
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

func TestEmptyTextsAreMistakes(t *testing.T) {
	for _, text := range []string{`""`, `{"en": ""}`, `null`} {
		world := strings.Replace(languageWorldJSON, `"Type who ({{.Step}}/{{.Steps}})."`, text, 1)
		_, err := LoadWorld(strings.NewReader(world))
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf(MsgErrTutorialStep, 2)) {
			t.Errorf(L("ett steg med texten %s har ingen text – väntade %q, fick %v",
				"a step with the text %s has no text – expected %q, got %v",
				"テキストが %s のステップにはテキストがありません。%q を期待しましたが %v"), text, fmt.Sprintf(MsgErrTutorialStep, 2), err)
		}
	}
}
