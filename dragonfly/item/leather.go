package item

// Leather does not have yet a Funtionality.
type Leather struct{}

// EncodeItem ...
func (Leather) EncodeItem() (id int32, meta int16) {
    return 334, 0
}
