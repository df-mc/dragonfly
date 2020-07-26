package item

// NetheriteIngot is a rare mineral crafted with 4 pieces of netherite scrap and 4 gold ingots.
type NetheriteIngot struct{}

// EncodeItem ...
func (NetheriteIngot) EncodeItem() (id int32, meta int16) {
	return 742, 0
}

// PayableForBeacon ...
func (NetheriteIngot) PayableForBeacon() bool {
	return true
}