package mud

import (
	"bytes"
	_ "embed" // needed for //go:embed
)

//go:embed world.json
var defaultWorldJSON []byte

// DefaultWorld is the world in world.json, baked into the program binary.
func DefaultWorld() *World {
	w, err := LoadWorld(bytes.NewReader(defaultWorldJSON))
	if err != nil {
		panic("world.json is broken: " + err.Error())
	}
	return w
}
