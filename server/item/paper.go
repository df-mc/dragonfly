package item

// Paper is an item crafted from sugar cane.
type Paper struct{}

func (Paper) EncodeItem() (name string, meta int16) {
	return "minecraft:paper", 0
}
