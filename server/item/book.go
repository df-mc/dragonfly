package item

// Book is an item used in enchanting and crafting.
type Book struct{}

// EncodeItem ...
func (Book) EncodeItem() (name string, meta int16) {
	return "minecraft:book", 0
}
