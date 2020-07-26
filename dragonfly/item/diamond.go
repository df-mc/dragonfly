package item

// Diamond is a rare mineral obtained from diamond ore or loot chests.
type Diamond struct{}

// EncodeItem ...
func (Diamond) EncodeItem() (id int32, meta int16) {
	return 264, 0
}

// PayableForBeacon ...
func (Diamond) PayableForBeacon() bool {
	return true
}
