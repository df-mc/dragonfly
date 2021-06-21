package item

// Bowls are containers that can hold certain foods.
type Bowl struct{}

// EncodeItem ...
func (Bowl) EncodeItem() (name string, meta int16) {
	return "minecraft:bowl", 0
}
