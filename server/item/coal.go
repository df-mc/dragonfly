package item

// Coal is an item used as fuel & crafting torches.
type Coal struct{} //TODO: Fuel

// EncodeItem ...
func (Coal) EncodeItem() (name string, meta int16) {
	return "minecraft:coal", 0
}
