package mud

import (
	"strings"
	"testing"
)

const smallWorld = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Torget", "description": "Ett torg.", "terrain": "town", "x": 0, "y": 0, "z": 0,
     "exits": {"north": "krog", "up": "torn"}},
    {"id": "krog", "name": "Krogen", "description": "En krog.", "terrain": "inside", "x": 0, "y": -1, "z": 0,
     "exits": {"south": "torg"}},
    {"id": "torn", "name": "Tornet", "description": "Högt upp.", "terrain": "inside", "x": 0, "y": 0, "z": 1,
     "exits": {"down": "torg"}}
  ]
}`

func TestLoadWorld(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(smallWorld))
	if err != nil {
		t.Fatalf("LoadWorld gav fel: %v", err)
	}
	if w.Start != "torg" || len(w.Rooms) != 3 {
		t.Fatalf("fick Start=%q och %d rum, väntade \"torg\" och 3", w.Start, len(w.Rooms))
	}
	krog := w.Room("krog")
	if krog == nil {
		t.Fatal("rummet krog saknas")
	}
	if krog.Name != "Krogen" || krog.Description != "En krog." || krog.Terrain != "inside" || krog.Y != -1 {
		t.Errorf("krog lästes in fel: %+v", *krog)
	}
	if w.Room("torg").Exits["up"] != "torn" {
		t.Errorf("torg ska ha utgången up -> torn, har %v", w.Room("torg").Exits)
	}
}

func TestLoadWorldErrors(t *testing.T) {
	room := func(id, extra string) string {
		return `{"id": "` + id + `", "name": "X", "description": "", "terrain": "town", ` + extra + `}`
	}
	tests := []struct {
		name string
		json string
		want string // felet ska innehålla den här texten
	}{
		{"trasig JSON", `{"start": "a", "rooms": [`, ""},
		{"stavfel i fältnamn", `{"start": "a", "rooms": [{"id": "a", "nmae": "A", "terrain": "town"}]}`, ""},
		{"startrummet saknas", `{"start": "b", "rooms": [` + room("a", `"x": 0`) + `]}`, `"b"`},
		{"utgång till okänt rum", `{"start": "a", "rooms": [` + room("a", `"exits": {"north": "spöke"}`) + `]}`, `"spöke"`},
		{"samma rum två gånger", `{"start": "a", "rooms": [` + room("a", `"x": 0`) + `, ` + room("a", `"x": 1`) + `]}`, `"a"`},
		{"okänd terräng", `{"start": "a", "rooms": [{"id": "a", "name": "A", "terrain": "lava"}]}`, `"lava"`},
		{"okänd riktning", `{"start": "a", "rooms": [` + room("a", `"exits": {"diagonalt": "a"}`) + `]}`, `"diagonalt"`},
		{"samma koordinater", `{"start": "a", "rooms": [` + room("a", `"x": 0`) + `, ` + room("b", `"x": 0`) + `]}`, "koordinat"},
		{"koordinaterna passar inte utgången", `{"start": "a", "rooms": [` +
			room("a", `"x": 0, "y": 0, "exits": {"north": "b"}`) + `, ` + room("b", `"x": 5, "y": 5`) + `]}`, "koordinat"},
	}
	for _, tt := range tests {
		_, err := LoadWorld(strings.NewReader(tt.json))
		if err == nil {
			t.Errorf("%s: LoadWorld ska ge fel", tt.name)
			continue
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: felet %q borde nämna %s", tt.name, err, tt.want)
		}
	}
}

func TestValidateReportsAllErrors(t *testing.T) {
	w := NewWorld("saknas",
		&Room{ID: "a", Name: "A", Terrain: "town", Exits: map[string]string{"north": "spöke"}},
	)
	err := w.Validate()
	if err == nil {
		t.Fatal("Validate ska hitta fel")
	}
	for _, want := range []string{`"saknas"`, `"spöke"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Validate ska rapportera ALLA fel; %q saknar %s", err, want)
		}
	}
}

func TestDefaultWorldFromFile(t *testing.T) {
	w := DefaultWorld()
	if err := w.Validate(); err != nil {
		t.Fatalf("standardvärlden har fel:\n%v", err)
	}
	if len(w.Rooms) < 10 {
		t.Errorf("standardvärlden har %d rum – minst 10 i world.json, tack!", len(w.Rooms))
	}
	terrains := map[string]bool{}
	for _, r := range w.Rooms {
		terrains[r.Terrain] = true
	}
	if len(terrains) < 4 {
		t.Errorf("använd minst 4 olika terrängtyper (har %v) – det blir snyggare på kartan i nästa lektion", terrains)
	}
}

func TestServerWithLoadedWorld(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(smallWorld))
	if err != nil {
		t.Fatal(err)
	}
	addr := startServer(t, NewServerWithWorld(w))
	c := login(t, addr, "Anna")
	c.expect("Torget")
	c.drain()
	c.send("up")
	c.expect("Tornet")
}
