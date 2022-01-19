package item

import (
	"time"
)

// Spyglass is an item that zooms in on an area the player is looking at, like a telescope.
type Spyglass struct{}

// MaxCount always returns 1.
func (Spyglass) MaxCount() int {
	return 1
}

// Release ...
func (Spyglass) Release(Releaser, time.Duration, *UseContext) {}

// Requirements ...
func (Spyglass) Requirements() []Stack {
	return []Stack{}
}

// EncodeItem ...
func (Spyglass) EncodeItem() (name string, meta int16) {
	return "minecraft:spyglass", 0
}
