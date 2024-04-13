package item

// Diamond is a rare mineral obtained from diamond ore or loot chests.
type Diamond struct{}

// EncodeItem ...
func (Diamond) EncodeItem() (name string, meta int16) {
	return "minecraft:diamond", 0
}

// TrimMaterial ...
func (Diamond) TrimMaterial() string {
	return "diamond"
}

// MaterialColor ...
func (Diamond) MaterialColor() string {
	return "s"
}

// PayableForBeacon ...
func (Diamond) PayableForBeacon() bool {
	return true
}
