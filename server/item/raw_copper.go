package item

// RawCopper is a raw metal resource obtained from mining copper ore.
type RawCopper struct{}

// EncodeItem ...
func (r RawCopper) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_copper", 0
}
