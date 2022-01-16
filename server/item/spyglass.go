package item

import (
	"time"
)

// SpyGlass is an item that zooms in on an area the player is looking at, like a telescope.
type SpyGlass struct{}

// MaxCount always returns 1.
func (SpyGlass) MaxCount() int {
	return 1
}

// Release ...
func (SpyGlass) Release(_ Releaser, _ time.Duration, _ *UseContext) {}

// Requirements ...
func (SpyGlass) Requirements() []Stack {
	return []Stack{}
}

// EncodeItem ...
func (SpyGlass) EncodeItem() (name string, meta int16) {
	return "minecraft:spyglass", 0
}
