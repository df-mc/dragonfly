package item

// GlisteringMelonSlice is an inedible item used for brewing potions of healing. It is also one of the many potion
// ingredients that can be used to make mundane potions.
type GlisteringMelonSlice struct{}

// EncodeItem ...
func (GlisteringMelonSlice) EncodeItem() (name string, meta int16) {
	return "minecraft:glistering_melon_slice", 0
}
