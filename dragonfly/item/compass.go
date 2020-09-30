package item

// The compass helps to find the spawn place.
type Compass struct{}

// EncodeItem ...
func (w Compass) EncodeItem() (id int32, meta int16) {
	return 345, 0
}
