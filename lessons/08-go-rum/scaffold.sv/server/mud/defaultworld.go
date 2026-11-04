package mud

// DefaultWorld is the village of Mudby. Feel free to add more rooms!
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
