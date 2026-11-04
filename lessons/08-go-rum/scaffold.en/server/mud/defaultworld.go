package mud

// DefaultWorld is the village of Mudby. Feel free to add more rooms!
func DefaultWorld() *World {
	return NewWorld("torget",
		&Room{
			ID:          "torget",
			Name:        "Mudby Square",
			Description: "A cobbled square with an old well in the middle. From the tavern\nto the north comes laughter and the clink of mugs.",
			Exits:       map[string]string{"north": "krogen", "east": "smedjan", "south": "porten", "west": "templet"},
		},
		&Room{
			ID:          "krogen",
			Name:        "The Merry Dragon Tavern",
			Description: "It smells of ale and smoke. A staircase leads up to the guest rooms.",
			Exits:       map[string]string{"south": "torget", "up": "gastrummet"},
		},
		&Room{
			ID:          "gastrummet",
			Name:        "A small guest room",
			Description: "A narrow bed and a window facing the square.",
			Exits:       map[string]string{"down": "krogen"},
		},
		&Room{
			ID:          "smedjan",
			Name:        "The Smithy",
			Description: "Heat from the forge hits you. Hammer blows echo off the walls.",
			Exits:       map[string]string{"west": "torget"},
		},
		&Room{
			ID:          "templet",
			Name:        "The Temple",
			Description: "A quiet, cool temple. Candles flicker before an altar.",
			Exits:       map[string]string{"east": "torget"},
		},
		&Room{
			ID:          "porten",
			Name:        "The Town Gate",
			Description: "A heavy wooden gate in the town wall. To the south a road winds towards the forest.",
			Exits:       map[string]string{"north": "torget"},
		},
	)
}
