package item

// Clock is used to measure and display in-game time.
type Clock struct{}

// EncodeItem ...
func (w Clock) EncodeItem() (name string, meta int16) {
	return "minecraft:clock", 0
}
