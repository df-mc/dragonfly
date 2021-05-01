package item

// Clock is used to measure and display in-game time.
type Clock struct{}

// EncodeItem ...
func (w Clock) EncodeItem() (id int32, name string, meta int16) {
	return 347, "minecraft:clock", 0
}
