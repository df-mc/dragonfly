package item

// IronIngot is a rare mineral melted from iron ore or obtained from loot chests.
type IronIngot struct{}

// EncodeItem ...
func (IronIngot) EncodeItem() (id int32, name string, meta int16) {
	return 265, "minecraft:iron_ingot", 0
}

// PayableForBeacon ...
func (IronIngot) PayableForBeacon() bool {
	return true
}
