package item

// Bone is an item primarily obtained as a drop from skeletons and their variants.
type Bone struct{}

// EncodeItem ...
func (Bone) EncodeItem() (name string, meta int16) {
	return "minecraft:bone", 0
}
