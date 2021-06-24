package item

import "github.com/df-mc/dragonfly/server/world"

// Salmon is a food item obtained from salmons. It can be cooked in a furnace, smoker, or campfire.
type Salmon struct {
	defaultFood

	// Cooked is whether the salmon is cooked.
	Cooked bool
}

// Consume ...
func (s Salmon) Consume(_ *world.World, c Consumer) Stack {
	if s.Cooked {
		c.Saturate(6, 9.6)
	} else {
		c.Saturate(2, 0.4)
	}
	return Stack{}
}

// EncodeItem ...
func (s Salmon) EncodeItem() (name string, meta int16) {
	if s.Cooked {
		return "minecraft:cooked_salmon", 0
	}
	return "minecraft:salmon", 0
}
