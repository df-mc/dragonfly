package item

// Magma cream is an item used in brewing to create potions of Fire Resistance, and to build magma blocks.
type MagmaCream struct{}

// EncodeItem ...
func (m MagmaCream) EncodeItem() (id int32, meta int16) {
	return 378, 0
}
