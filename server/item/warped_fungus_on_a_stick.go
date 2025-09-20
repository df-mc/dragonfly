package item

// WarpedFungusOnAStick is an item that can be used to control saddled striders.
type WarpedFungusOnAStick struct{}

func (WarpedFungusOnAStick) MaxCount() int {
	return 1
}

func (WarpedFungusOnAStick) EncodeItem() (name string, meta int16) {
	return "minecraft:warped_fungus_on_a_stick", 0
}
