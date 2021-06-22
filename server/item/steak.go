package item

import "github.com/df-mc/dragonfly/server/world"

// Steak is a food item obtained from cows or from cooking raw beef.
type Steak struct {
	defaultFood
}

// Consume ...
func (Steak) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(8, 12.8)
	return Stack{}
}

// EncodeItem ...
func (Steak) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_beef", 0
}
