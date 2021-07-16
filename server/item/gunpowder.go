package item

// Gunpowder is an item that is used for explosion-related recipes.
type Gunpowder struct{}

// EncodeItem ...
func (Gunpowder) EncodeItem() (name string, meta int16) {
	return "minecraft:gunpowder", 0
}
