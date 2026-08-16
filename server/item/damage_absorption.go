package item

// DamageAbsorption represents an item that can absorb damage that would otherwise be dealt to its
// wearer. The item needs a minecraft:durability component for this to function.
type DamageAbsorption interface {
	// AbsorbableCauses returns the list of damage causes that can be absorbed by the item.
	AbsorbableCauses() []string
}
