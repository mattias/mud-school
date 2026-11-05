package mud

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

const itemWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "exits": {"north": "krog"}},
    {"id": "krog", "name": "Tavern", "description": "A tavern.", "terrain": "inside", "y": -1, "exits": {"south": "torg"}}
  ],
  "items": [
    {"id": "sword", "name": "a rusty sword", "description": "Old but sharp.", "keywords": ["sword", "rusty"], "room": "torg"},
    {"id": "mug", "name": "a clay mug", "description": "It smells of ale.", "keywords": ["mug", "cup"], "room": "torg"},
    {"id": "bread", "name": "a loaf of bread", "description": "Still warm.", "keywords": ["bread"], "room": "krog"}
  ]
}`

func itemWorld(t *testing.T) *World {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(itemWorldJSON))
	if err != nil {
		t.Fatalf(L("världen med föremål gick inte att läsa: %v", "the world with items could not be loaded: %v", "アイテム付きの世界を読み込めませんでした: %v"), err)
	}
	return w
}

type word string

func (w word) Matches(s string) bool { return strings.EqualFold(string(w), s) }

func TestFindMatch(t *testing.T) {
	list := []word{"apple", "Banana", "cherry"}
	if got, ok := FindMatch(list, "banana"); !ok || got != "Banana" {
		t.Errorf(L("FindMatch(..., %q) = %q, %v; väntade %q, true", "FindMatch(..., %q) = %q, %v; expected %q, true", "FindMatch(..., %q) = %q, %v ですが、%q, true を期待しました"), "banana", got, ok, "Banana")
	}
	if got, ok := FindMatch(list, "kiwi"); ok || got != "" {
		t.Errorf(L("FindMatch för något som saknas ska ge nollvärdet och false, gav %q, %v", "FindMatch for something missing should return the zero value and false, gave %q, %v", "見つからない場合 FindMatch はゼロ値と false を返すべきですが、%q, %v でした"), got, ok)
	}
	sword := &Item{ID: "sword", Name: "a sword", Keywords: []string{"blade", "Rusty"}}
	for _, w := range []string{"sword", "SWORD", "blade", "rusty"} {
		if !sword.Matches(w) {
			t.Errorf(L("föremålet ska matcha %q (id eller keyword, oavsett skiftläge)", "the item should match %q (id or keyword, any letter case)", "アイテムは %q に一致するべきです（id かキーワード、大文字小文字は無視）"), w)
		}
	}
	if sword.Matches("a") {
		t.Error(L("föremålet ska inte matcha delar av ord", "the item should not match parts of words", "アイテムは単語の一部には一致しないべきです"))
	}
}

func TestLoadItemsErrors(t *testing.T) {
	tests := []struct{ name, items, want string }{
		{"missing room", `{"items": [{"id": "x", "name": "X", "room": "nowhere"}]}`, `"nowhere"`},
		{"twice", `{"items": [{"id": "x", "name": "X", "room": "torg"}, {"id": "x", "name": "Y", "room": "torg"}]}`, `"x"`},
		{"no name", `{"items": [{"id": "x", "room": "torg"}]}`, `"x"`},
		{"typo", `{"items": [{"id": "x", "nmae": "X", "room": "torg"}]}`, ""},
	}
	for _, tt := range tests {
		w := itemWorld(t)
		err := w.LoadItems(strings.NewReader(tt.items))
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf(L("%s: LoadItems ska ge ett fel som nämner %s, gav %v", "%s: LoadItems should return an error mentioning %s, returned %v", "%s: LoadItems は %s を含むエラーを返すべきですが、%v でした"), tt.name, tt.want, err)
		}
	}
	w := itemWorld(t)
	if err := w.LoadItems(strings.NewReader(`{"items": [{"id": "coin", "name": "a coin", "room": "krog"}]}`)); err != nil {
		t.Fatal(err)
	}
	if w.Items["coin"] == nil || len(w.Room("krog").Items) != 2 {
		t.Error(L("LoadItems ska lägga föremålet i World.Items och i rummet", "LoadItems should put the item in World.Items and in its room", "LoadItems はアイテムを World.Items と部屋に入れるべきです"))
	}
}

func TestDefaultWorldHasItems(t *testing.T) {
	w := DefaultWorld()
	if len(w.Items) < 5 {
		t.Errorf(L("standardvärlden har %d föremål – minst 5 i items.json, tack!", "the default world has %d items – at least 5 in items.json, please!", "標準の世界のアイテムは %d 個です。items.json には 5 個以上入れてください！"), len(w.Items))
	}
}

func TestLookShowsItems(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(itemWorld(t)))
	c := login(t, addr, "Anna")
	c.expect("Square")
	c.expect("A square.")
	c.expectContains(strings.SplitN(MsgExits, "%s", 2)[0])
	c.expect(fmt.Sprintf(MsgItemsHere, "a rusty sword"+MsgListSep+"a clay mug"))
}

func TestGetDropInventory(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(itemWorld(t)))
	anna := login(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	anna.drain()
	bertil.drain()

	anna.send("inventory")
	anna.expect(MsgInventoryEmpty)
	anna.send("get")
	anna.expect(MsgGetWhat)
	anna.send("get dragon")
	anna.expect(fmt.Sprintf(MsgNoSuchItem, "dragon"))

	anna.send("get rusty")
	anna.expect(fmt.Sprintf(MsgYouGet, "a rusty sword"))
	bertil.expect(fmt.Sprintf(MsgGets, "Anna", "a rusty sword"))

	anna.send("take cup")
	anna.expect(fmt.Sprintf(MsgYouGet, "a clay mug"))
	bertil.expect(fmt.Sprintf(MsgGets, "Anna", "a clay mug"))

	anna.send("i")
	anna.expect(fmt.Sprintf(MsgInventory, "a rusty sword"+MsgListSep+"a clay mug"))

	bertil.send("look")
	bertil.expect("Square")
	bertil.expect("A square.")
	bertil.expectContains(strings.SplitN(MsgExits, "%s", 2)[0])
	if line, _ := bertil.readLine(testTimeout); strings.Contains(line, "sword") {
		t.Errorf(L("svärdet ska inte längre synas i rummet: %q", "the sword should no longer be in the room: %q", "剣はもう部屋に見えないはずです: %q"), line)
	}

	anna.send("drop")
	anna.expect(MsgDropWhat)
	anna.send("drop bread")
	anna.expect(fmt.Sprintf(MsgNotCarrying, "bread"))
	anna.send("drop sword")
	anna.expect(fmt.Sprintf(MsgYouDrop, "a rusty sword"))
	bertil.drain()

	anna.send("n")
	anna.drain()
	anna.send("drop mug")
	anna.expect(fmt.Sprintf(MsgYouDrop, "a clay mug"))
	anna.send("look")
	anna.expect("Tavern")
	anna.expect("A tavern.")
	anna.expectContains(strings.SplitN(MsgExits, "%s", 2)[0])
	anna.expect(fmt.Sprintf(MsgItemsHere, "a loaf of bread"+MsgListSep+"a clay mug"))
}

func TestExamine(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(itemWorld(t)))
	c := login(t, addr, "Anna")
	c.drain()
	c.send("examine")
	c.expect(MsgExamineWhat)
	c.send("x sword")
	c.expect("Old but sharp.")
	c.send("get mug")
	c.drain()
	c.send("examine cup")
	c.expect("It smells of ale.")
	c.send("x bread")
	c.expect(fmt.Sprintf(MsgNoSuchItem, "bread"))
}

func TestTextKinds(t *testing.T) {
	addr := startServer(t, NewServerWithWorld(itemWorld(t)))
	anna := jsonLoginKinds(t, addr, "Anna")
	bertil := login(t, addr, "Bertil")
	bertil.drain()
	anna.drainFrames() // "Bertil has connected."

	expectKind := func(want string) {
		t.Helper()
		deadline := time.Now().Add(testTimeout)
		for time.Now().Before(deadline) {
			f := anna.readFrame()
			if f["type"] != "text" {
				continue
			}
			if f["kind"] != want {
				t.Errorf(L("text-framen %q ska ha \"kind\": %q, har %v", "the text frame %q should have \"kind\": %q, has %v", "テキストフレーム %q の \"kind\" は %q であるべきですが、%v でした"), f["text"], want, f["kind"])
			}
			return
		}
	}
	anna.send("look")
	expectKind(KindRoom)
	anna.drainFrames()
	bertil.send("say hi")
	expectKind(KindSay)
	anna.send("fly")
	expectKind(KindError)
	anna.send("west")
	expectKind(KindError)
	bertil.send("n")
	expectKind(KindMove)
}

// jsonLoginKinds logs in with the JSON protocol and reads away the frames after login.
func jsonLoginKinds(t *testing.T, addr, name string) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expect(Welcome)
	c.expect(MsgAskName)
	c.send("@hello mudschool 1")
	c.send(name)
	c.drainFrames()
	return c
}

func (c *testClient) drainFrames() { c.drain() }
