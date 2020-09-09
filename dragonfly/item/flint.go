package item

// Flint is an item dropped rarely by gravel.
type Flint struct{}

// EncodeItem ...
func (f Flint) EncodeItem() (id int32, meta int16) {
	return 318, 0
}
