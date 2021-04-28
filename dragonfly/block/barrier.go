package block

// Barrier is a transparent solid block used to create invisible boundaries.
type Barrier struct {
	transparent
	solid
}

// EncodeItem ...
func (Barrier) EncodeItem() (id int32, meta int16) {
	return -161, 0
}

// EncodeBlock ...
func (Barrier) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:barrier", nil
}
