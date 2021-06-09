package item

// IronIngot is a metal ingot melted from raw iron or obtained from loot chests.
type IronIngot struct{}

// EncodeItem ...
func (IronIngot) EncodeItem() (name string, meta int16) {
	return "minecraft:iron_ingot", 0
}

// PayableForBeacon ...
func (IronIngot) PayableForBeacon() bool {
	return true
}
