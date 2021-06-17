package item

// RawGold is a raw metal resource obtained from mining gold ore.
type RawGold struct{}

// EncodeItem ...
func (r RawGold) EncodeItem() (name string, meta int16) {
	return "minecraft:raw_gold", 0
}
