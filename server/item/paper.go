package item

// Paper is an item crafted from sugar cane.
type Paper struct{}

// EncodeItem ...
func (Paper) EncodeItem() (name string, meta int16) {
	return "minecraft:paper", 0
}
