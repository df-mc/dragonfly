package item

import "github.com/df-mc/dragonfly/server/world"

// CookedMutton is a food item obtained from cooking raw mutton.
type CookedMutton struct {
	defaultFood
}

// Consume ...
func (CookedMutton) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 9.6)
	return Stack{}
}

// EncodeItem ...
func (CookedMutton) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_mutton", 0
}
