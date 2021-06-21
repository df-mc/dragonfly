package item

// NetherBrick is an item made by smelting netherrack in a furnace.
type NetherBrick struct{}

// EncodeItem ...
func (NetherBrick) EncodeItem() (name string, meta int16) {
	return "minecraft:netherbrick", 0
}
