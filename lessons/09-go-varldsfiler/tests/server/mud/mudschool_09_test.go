package mud

import (
	"strings"
	"testing"
)

const smallWorld = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "x": 0, "y": 0, "z": 0,
     "exits": {"north": "krog", "up": "torn"}},
    {"id": "krog", "name": "Tavern", "description": "A tavern.", "terrain": "inside", "x": 0, "y": -1, "z": 0,
     "exits": {"south": "torg"}},
    {"id": "torn", "name": "Tower", "description": "High up.", "terrain": "inside", "x": 0, "y": 0, "z": 1,
     "exits": {"down": "torg"}}
  ]
}`

func TestLoadWorld(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(smallWorld))
	if err != nil {
		t.Fatalf(L("LoadWorld gav fel: %v", "LoadWorld returned an error: %v", "LoadWorld がエラーを返しました: %v"), err)
	}
	if w.Start != "torg" || len(w.Rooms) != 3 {
		t.Fatalf(L("fick Start=%q och %d rum, väntade \"torg\" och 3",
			"got Start=%q and %d rooms, expected \"torg\" and 3",
			"Start=%q、部屋 %d 個でした。\"torg\" と 3 個を期待しました"), w.Start, len(w.Rooms))
	}
	krog := w.Room("krog")
	if krog == nil {
		t.Fatal(L("rummet krog saknas", "the room krog is missing", "部屋 krog がありません"))
	}
	if krog.Name != "Tavern" || krog.Description != "A tavern." || krog.Terrain != "inside" || krog.Y != -1 {
		t.Errorf(L("krog lästes in fel: %+v", "krog was read incorrectly: %+v", "krog の読み込み結果が違います: %+v"), *krog)
	}
	if w.Room("torg").Exits["up"] != "torn" {
		t.Errorf(L("torg ska ha utgången up -> torn, har %v", "torg should have the exit up -> torn, has %v", "torg には出口 up -> torn が必要ですが、%v でした"), w.Room("torg").Exits)
	}
}

func TestLoadWorldErrors(t *testing.T) {
	room := func(id, extra string) string {
		return `{"id": "` + id + `", "name": "X", "description": "", "terrain": "town", ` + extra + `}`
	}
	tests := []struct {
		name string
		json string
		want string // the error must contain this text
	}{
		{"broken JSON", `{"start": "a", "rooms": [`, ""},
		{"typo in a field name", `{"start": "a", "rooms": [{"id": "a", "nmae": "A", "terrain": "town"}]}`, ""},
		{"start room missing", `{"start": "b", "rooms": [` + room("a", `"x": 0`) + `]}`, `"b"`},
		{"exit to an unknown room", `{"start": "a", "rooms": [` + room("a", `"exits": {"north": "ghost"}`) + `]}`, `"ghost"`},
		{"same room twice", `{"start": "a", "rooms": [` + room("a", `"x": 0`) + `, ` + room("a", `"x": 1`) + `]}`, `"a"`},
		{"unknown terrain", `{"start": "a", "rooms": [{"id": "a", "name": "A", "terrain": "lava"}]}`, `"lava"`},
		{"unknown direction", `{"start": "a", "rooms": [` + room("a", `"exits": {"diagonal": "a"}`) + `]}`, `"diagonal"`},
		{"same coordinates", `{"start": "a", "rooms": [` + room("a", `"x": 0`) + `, ` + room("b", `"x": 0`) + `]}`, `"b"`},
		{"coordinates do not match the exit", `{"start": "a", "rooms": [` +
			room("a", `"x": 0, "y": 0, "exits": {"north": "b"}`) + `, ` + room("b", `"x": 5, "y": 5`) + `]}`, `"b"`},
	}
	for _, tt := range tests {
		_, err := LoadWorld(strings.NewReader(tt.json))
		if err == nil {
			t.Errorf(L("%s: LoadWorld ska ge fel", "%s: LoadWorld should return an error", "%s: LoadWorld はエラーを返すべきです"), tt.name)
			continue
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Errorf(L("%s: felet %q borde nämna %s", "%s: the error %q should mention %s", "%s: エラー %q には %s が含まれるべきです"), tt.name, err, tt.want)
		}
	}
}

func TestValidateReportsAllErrors(t *testing.T) {
	w := NewWorld("missing",
		&Room{ID: "a", Name: "A", Terrain: "town", Exits: map[string]string{"north": "ghost"}},
	)
	err := w.Validate()
	if err == nil {
		t.Fatal(L("Validate ska hitta fel", "Validate should find errors", "Validate はエラーを見つけるべきです"))
	}
	for _, want := range []string{`"missing"`, `"ghost"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf(L("Validate ska rapportera ALLA fel; %q saknar %s",
				"Validate should report ALL errors; %q is missing %s",
				"Validate はすべてのエラーを報告するべきです。%q に %s がありません"), err, want)
		}
	}
}

func TestDefaultWorldFromFile(t *testing.T) {
	w := DefaultWorld()
	if err := w.Validate(); err != nil {
		t.Fatalf(L("standardvärlden har fel:\n%v", "the default world has errors:\n%v", "標準の世界にエラーがあります:\n%v"), err)
	}
	if len(w.Rooms) < 10 {
		t.Errorf(L("standardvärlden har %d rum – minst 10 i world.json, tack!",
			"the default world has %d rooms – at least 10 in world.json, please!",
			"標準の世界の部屋は %d 個です。world.json には 10 個以上入れてください！"), len(w.Rooms))
	}
	terrains := map[string]bool{}
	for _, r := range w.Rooms {
		terrains[r.Terrain] = true
	}
	if len(terrains) < 4 {
		t.Errorf(L("använd minst 4 olika terrängtyper (har %v) – det blir snyggare på kartan i nästa lektion",
			"use at least 4 different terrains (you have %v) – the map in the next lesson will look nicer",
			"地形は 4 種類以上使ってください（現在 %v）。次のレッスンのマップがきれいになります"), terrains)
	}
}

func TestServerWithLoadedWorld(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(smallWorld))
	if err != nil {
		t.Fatal(err)
	}
	addr := startServer(t, NewServerWithWorld(w))
	c := login(t, addr, "Anna")
	c.expect("Square")
	c.drain()
	c.send("up")
	c.expect("Tower")
}
