package item

// Raft is an item used to sail across water. Bamboo wood uses a raft instead of a boat.
// Unlike regular wood boats, bamboo rafts have a flat design.
type Raft struct {
	// Chest specifies whether the raft has a chest in it.
	Chest bool
}

// EncodeItem ...
func (r Raft) EncodeItem() (name string, meta int16) {
	if r.Chest {
		return "minecraft:bamboo_chest_raft", 0
	}
	return "minecraft:bamboo_raft", 0
}

// MaxCount ...
func (Raft) MaxCount() int {
	return 1
}
