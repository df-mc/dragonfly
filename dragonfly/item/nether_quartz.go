package item

// NetherQuartz is a smooth, white mineral found in the Nether.
type NetherQuartz struct{}

// EncodeItem ...
func (NetherQuartz) EncodeItem() (id int32, meta int16) {
	return 406, 0
}
