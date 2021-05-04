package item

// Leather is an animal skin used to make item frames, armor and books.
type Leather struct{}

// EncodeItem ...
func (Leather) EncodeItem() (name string, meta int16) {
	return "minecraft:leather", 0
}
