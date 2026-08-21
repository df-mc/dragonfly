package item

// Dyeable represents an item that can be dyed using dyes in a crafting grid, like leather armor.
type Dyeable interface {
	// DefaultColor returns the default RGB color of the item when undyed.
	DefaultColor() [3]uint8
}
