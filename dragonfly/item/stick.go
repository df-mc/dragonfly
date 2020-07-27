package item

// Sticks are one of the most abundant resources used for crafting many tools and items.
type Stick struct {}

func (s Stick) EncodeItem() (id int32, meta int16) {
	return 280, 0
}