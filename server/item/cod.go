package item

import "github.com/df-mc/dragonfly/server/world"

// Cod is a food item obtained from cod. It can be cooked in a furnace, smoker, or campfire.
type Cod struct {
	defaultFood

	// Cooked is whether the cod is cooked.
	Cooked bool
}

// Consume ...
func (co Cod) Consume(_ *world.World, c Consumer) Stack {
	if co.Cooked {
		c.Saturate(5, 6)
	} else {
		c.Saturate(2, 0.4)
	}
	return Stack{}
}

// EncodeItem ...
func (co Cod) EncodeItem() (name string, meta int16) {
	if co.Cooked {
		return "minecraft:cooked_cod", 0
	}
	return "minecraft:cod", 0
}
