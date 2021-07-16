package item

// MagmaCream is an item used in brewing to create potions of Fire Resistance, and to build magma blocks.
type MagmaCream struct{}

// EncodeItem ...
func (m MagmaCream) EncodeItem() (name string, meta int16) {
	return "minecraft:magma_cream", 0
}
