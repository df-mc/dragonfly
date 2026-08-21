package item

// EnchantableData represents an item that can be enchanted and provides the data required for the client to
// display the enchantability of the item.
type EnchantableData interface {
	// EnchantableData returns the enchantable data of the item.
	EnchantableData() EnchantableInfo
}

// EnchantableInfo is a struct returned by items that implement EnchantableData. It contains the information
// required for the client to display the enchantability of the item.
type EnchantableInfo struct {
	// Slot is the enchantment slot of the item, such as "melee_spear" or "sword".
	Slot string
	// Value is the enchantment value of the item.
	Value uint
}
