package item

// FireResistant represents an item that is resistant to fire and lava, such as netherite items.
type FireResistant interface {
	// FireResistant returns whether the item is resistant to fire and lava.
	FireResistant() bool
}
