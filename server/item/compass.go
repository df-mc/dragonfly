package item

// Compass is an item used to find the spawn position of a world.
type Compass struct{}

// EncodeItem ...
func (Compass) EncodeItem() (name string, meta int16) {
	return "minecraft:compass", 0
}
