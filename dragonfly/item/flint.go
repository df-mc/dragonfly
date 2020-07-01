package item

// Flint is a mineral obtained from gravel.
type Flint struct {
}

// EncodeItem ...
func (f Flint) EncodeItem() (id int32, meta int16) {
	return 318, 0
}
