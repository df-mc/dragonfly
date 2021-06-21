package item

// Honeycomb is an item obtained from bee nests and beehives.
type Honeycomb struct{}

// EncodeItem ...
func (Honeycomb) EncodeItem() (name string, meta int16) {
	return "minecraft:honeycomb", 0
}
