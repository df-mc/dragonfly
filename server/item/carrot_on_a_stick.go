package item

// CarrotOnAStick is an item that can be used to control saddled pigs.
type CarrotOnAStick struct{}

// MaxCount ...
func (CarrotOnAStick) MaxCount() int {
	return 1
}

// EncodeItem ...
func (CarrotOnAStick) EncodeItem() (name string, meta int16) {
	return "minecraft:carrot_on_a_stick", 0
}
