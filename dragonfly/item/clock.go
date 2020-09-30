package item

// Clocks are used to measure and display in-game time.
type Clock struct{}

// EncodeItem ...
func (w Clock) EncodeItem() (id int32, meta int16) {
	return 347, 0
}
