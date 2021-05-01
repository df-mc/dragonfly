package item

// GoldIngot is a rare mineral melted from golden ore or obtained from loot chests.
type GoldIngot struct{}

// EncodeItem ...
func (GoldIngot) EncodeItem() (id int32, name string, meta int16) {
	return 266, "minecraft:gold_ingot", 0
}

// PayableForBeacon ...
func (GoldIngot) PayableForBeacon() bool {
	return true
}
