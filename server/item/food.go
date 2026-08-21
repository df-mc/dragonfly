package item

// Food represents an item that has nutritional value and provides the data required for the client to
// display the nutritional value of the item.
type Food interface {
	// FoodInfo returns the food information of the item.
	FoodInfo() FoodInfo
}

// FoodInfo is a struct returned by items that implement Food. It contains the information required for the
// client to display the nutritional value of the item.
type FoodInfo struct {
	// Nutrition is the number of hunger points the item restores.
	Nutrition int
	// SaturationModifier is the modifier applied to the saturation restored by the item.
	SaturationModifier float64
	// UsingConvertsTo is the identifier of the item the item converts to when consumed, such as a bowl or
	// glass bottle.
	UsingConvertsTo string
	// OnUseAction is the action performed when the item is used, such as the chorus fruit teleport.
	OnUseAction int
	// OnUseRange is the range in blocks the on use action applies to.
	OnUseRange [3]float64
	// CooldownTime is the duration in seconds of the cooldown applied when the item is consumed.
	CooldownTime int
	// CooldownType is the type of cooldown applied when the item is consumed.
	CooldownType string
	// Effects is a list of effects applied to the consumer when the item is consumed.
	Effects []FoodEffect
	// RemoveEffects is a list of effect IDs removed from the consumer when the item is consumed.
	RemoveEffects []int
}

// FoodEffect is a struct returned by items that implement Food. It contains the information required for the
// client to display an effect applied when the item is consumed.
type FoodEffect struct {
	// ID is the ID of the effect.
	ID int
	// Duration is the duration in seconds of the effect.
	Duration int
	// Amplifier is the amplifier of the effect.
	Amplifier int
	// Chance is the chance the effect is applied.
	Chance float64
	// Name is the name of the effect, such as "hunger" or "poison".
	Name string
}
