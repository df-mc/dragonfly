package item

// GoldIngot is a rare mineral melted from golden ore or obtained from loot chests.
type GoldIngot struct{}

// EncodeItem ...
func (GoldIngot) EncodeItem() (id int32, meta int16) {
	return 266, 0
}

// PayableForBeacon ...
func (GoldIngot) PayableForBeacon() bool {
	return true
}