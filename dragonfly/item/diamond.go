package item

// Diamond is a rare mineral obtained from diamond ore or loot chests.
type Diamond struct{}

// EncodeItem ...
func (Diamond) EncodeItem() (id int32, name string, meta int16) {
	return 264, "minecraft:diamond", 0
}

// PayableForBeacon ...
func (Diamond) PayableForBeacon() bool {
	return true
}
