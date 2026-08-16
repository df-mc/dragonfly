package item

// Rarity represents an item with a specific base rarity that determines the color of the item name
// when hovering over it.
type Rarity interface {
	// Rarity returns the base rarity of the item. Valid values are "common", "uncommon", "rare",
	// and "epic".
	Rarity() string
}
