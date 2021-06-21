package item

// Books are items used in enchanting and crafting.
type Book struct{}

// EncodeItem ...
func (Book) EncodeItem() (name string, meta int16) {
	return "minecraft:book", 0
}
