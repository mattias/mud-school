package mud

import (
	"bytes"
	_ "embed" // needed for //go:embed
)

//go:embed world.json
var defaultWorldJSON []byte

//go:embed items.json
var defaultItemsJSON []byte

// DefaultWorld is the world in world.json and items.json, baked into the
// program binary.
func DefaultWorld() *World {
	w, err := LoadWorld(bytes.NewReader(defaultWorldJSON))
	if err != nil {
		panic("world.json is broken: " + err.Error())
	}
	if err := w.LoadItems(bytes.NewReader(defaultItemsJSON)); err != nil {
		panic("items.json is broken: " + err.Error())
	}
	return w
}
