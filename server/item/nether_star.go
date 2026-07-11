package item

// NetherStar is a rare item dropped by the wither that is used solely to craft beacons.
type NetherStar struct{}

// EncodeItem ...
func (NetherStar) EncodeItem() (name string, meta int16) {
	return "minecraft:nether_star", 0
}

// BlastProof indicates that the item will withstand an explosion as an item entity.
func (NetherStar) BlastProof() bool {
	return true
}
