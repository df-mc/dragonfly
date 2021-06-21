package item

// Sugar is a food ingredient and brewing ingredient made from sugar canes.
type Sugar struct{}

// EncodeItem ...
func (Sugar) EncodeItem() (name string, meta int16) {
	return "minecraft:sugar", 0
}
