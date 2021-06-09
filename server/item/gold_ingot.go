package item

// GoldIngot is a metal ingot melted from raw gold or obtained from loot chests.
type GoldIngot struct{}

// EncodeItem ...
func (GoldIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:gold_ingot", 0
}

// PayableForBeacon ...
func (GoldIngot) PayableForBeacon() bool {
	return true
}
