package item

// Charcoal is an item obtained by smelting logs or wood.
type Charcoal struct{}

// EncodeItem ...
func (Charcoal) EncodeItem() (name string, meta int16) {
	return "minecraft:charcoal", 0
}
