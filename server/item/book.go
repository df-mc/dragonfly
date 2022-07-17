package item

// Book is an item used in enchanting and crafting.
type Book struct{}

// EnchantmentValue ...
func (b Book) EnchantmentValue() int {
	return 1
}

// EncodeItem ...
func (Book) EncodeItem() (name string, meta int16) {
	return "minecraft:book", 0
}
