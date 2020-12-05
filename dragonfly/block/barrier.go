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
