package item

// EnchantedBook is an item that lets players add enchantments to certain items using an anvil.
type EnchantedBook struct{}

// EncodeItem ...
func (EnchantedBook) EncodeItem() (name string, meta int16) {
	return "minecraft:enchanted_book", 0
}
