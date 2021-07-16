package item

// Stick is one of the most abundant resources used for crafting many tools and items.
type Stick struct{}

// EncodeItem ...
func (s Stick) EncodeItem() (name string, meta int16) {
	return "minecraft:stick", 0
}
