package item

// HeartOfTheSea is a rare item that can be crafted into a conduit.
type HeartOfTheSea struct{}

// EncodeItem ...
func (HeartOfTheSea) EncodeItem() (name string, meta int16) {
	return "minecraft:heart_of_the_sea", 0
}
