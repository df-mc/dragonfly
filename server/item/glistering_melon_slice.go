package item

// GlisteringMelonSlice is an inedible item used for brewing potions of healing.
type GlisteringMelonSlice struct{}

// EncodeItem ...
func (GlisteringMelonSlice) EncodeItem() (name string, meta int16) {
	return "minecraft:glistering_melon_slice", 0
}
