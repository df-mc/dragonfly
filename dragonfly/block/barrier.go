package block

// Barrier is a transparent solid block used to create invisible boundaries.
type Barrier struct {
	noNBT
	transparent
	solid
}

// EncodeItem ...
func (Barrier) EncodeItem() (id int32, meta int16) {
	return -161, 0
}

// EncodeBlock ...
func (Barrier) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:barrier", nil
}

// Hash ...
func (Barrier) Hash() uint64 {
	return hashBarrier
}
