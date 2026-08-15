package item

// UseModifiers represents an item that modifies the way it is used, such as slowing down the user while it is
// being used.
type UseModifiers interface {
	// UseModifiers returns the use modifiers of the item.
	UseModifiers() UseModifiersInfo
}

// UseModifiersInfo is a struct returned by items that implement UseModifiers. It contains the information
// required for the client to modify the way the item is used.
type UseModifiersInfo struct {
	// MovementModifier is the modifier applied to the movement speed of the user while using the item.
	MovementModifier float64
	// UseDuration is the duration in seconds the item takes to be fully used.
	UseDuration float64
	// EmitVibrations is true if using the item emits vibrations.
	EmitVibrations bool
	// StartSound is the sound played when the item starts being used.
	StartSound string
	// StartUsing is the condition required to start using the item, such as "always" or "require_charging".
	StartUsing string
}
