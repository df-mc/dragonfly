package item

// Arrow is used as ammunition for bows, crossbows, and dispensers. Arrows can be modified to
// imbue status effects on players and mobs.
type Arrow struct{}

// EncodeItem ...
func (Arrow) EncodeItem() (name string, meta int16) {
	return "minecraft:arrow", 0
}
