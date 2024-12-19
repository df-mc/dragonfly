package item

// Spyglass is an item that zooms in on an area the player is looking at, like a telescope.
type Spyglass struct {
	nopReleasable
}

// MaxCount always returns 1.
func (Spyglass) MaxCount() int {
	return 1
}

// EncodeItem ...
func (Spyglass) EncodeItem() (name string, meta int16) {
	return "minecraft:spyglass", 0
}
