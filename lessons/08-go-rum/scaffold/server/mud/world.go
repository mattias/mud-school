package mud

// Room är en plats i världen.
type Room struct {
	ID          string
	Name        string
	Description string
	Exits       map[string]string // riktning ("north") -> rummets ID
}

// World är alla rum i spelet.
type World struct {
	Start string           // ID för rummet där nya spelare hamnar
	Rooms map[string]*Room // ID -> rum
}

// Directions är alla riktningar i den ordning de ska visas.
var Directions = []string{"north", "east", "south", "west", "up", "down"}

// directionNames är riktningarnas svenska namn, som de visas för spelaren.
var directionNames = map[string]string{
	"north": "norr", "east": "öster", "south": "söder", "west": "väster", "up": "upp", "down": "ner",
}

// directionTravel används när någon går: "Anna går norrut."
var directionTravel = map[string]string{
	"north": "norrut", "east": "österut", "south": "söderut", "west": "västerut", "up": "uppåt", "down": "nedåt",
}

// NewWorld skapar en värld av rummen. "..." gör funktionen variadisk:
// NewWorld("torg", rum1, rum2, rum3) – rooms blir en []*Room.
func NewWorld(start string, rooms ...*Room) *World {
	// TODO(lektion 8): skapa en World med en map från ID till rum.
	return nil
}

// Room returnerar rummet med givet ID, eller nil om det inte finns.
func (w *World) Room(id string) *Room {
	// TODO(lektion 8)
	return nil
}

// DefaultWorld är byn Mudby. Hitta gärna på fler rum!
func DefaultWorld() *World {
	return NewWorld("torget",
		&Room{
			ID:          "torget",
			Name:        "Torget i Mudby",
			Description: "Ett kullerstensbelagt torg. En gammal brunn står i mitten och\nfrån krogen i norr hörs skratt och klirr.",
			Exits:       map[string]string{"north": "krogen", "east": "smedjan", "south": "porten", "west": "templet"},
		},
		&Room{
			ID:          "krogen",
			Name:        "Krogen Den Glada Draken",
			Description: "Det luktar öl och rök. En trappa leder upp till gästrummen.",
			Exits:       map[string]string{"south": "torget", "up": "gastrummet"},
		},
		&Room{
			ID:          "gastrummet",
			Name:        "Ett litet gästrum",
			Description: "En smal säng och ett fönster mot torget.",
			Exits:       map[string]string{"down": "krogen"},
		},
		&Room{
			ID:          "smedjan",
			Name:        "Smedjan",
			Description: "Värmen från ässjan slår emot dig. Hammarslag ekar mellan väggarna.",
			Exits:       map[string]string{"west": "torget"},
		},
		&Room{
			ID:          "templet",
			Name:        "Templet",
			Description: "Ett tyst, svalt tempel. Ljus fladdrar framför ett altare.",
			Exits:       map[string]string{"east": "torget"},
		},
		&Room{
			ID:          "porten",
			Name:        "Stadsporten",
			Description: "En tung träport i stadsmuren. Söderut slingrar sig en väg mot skogen.",
			Exits:       map[string]string{"north": "torget"},
		},
	)
}
