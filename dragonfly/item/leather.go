package item

// Leather is an animal skin used to make item frames, armor and books.
type Leather struct{}

// EncodeItem ...
func (Leather) EncodeItem() (id int32, meta int16) {
	return 334, 0
}
