package item

import "github.com/df-mc/dragonfly/server/world"

// Porkchop is a food item obtained from pigs. It can be cooked in a furnace, smoker, or campfire.
type Porkchop struct {
	defaultFood

	// Cooked is whether the porkchop is cooked.
	Cooked bool
}

// Consume ...
func (p Porkchop) Consume(_ *world.World, c Consumer) Stack {
	if p.Cooked {
		c.Saturate(8, 12.8)
	} else {
		c.Saturate(3, 1.8)
	}
	return Stack{}
}

// EncodeItem ...
func (p Porkchop) EncodeItem() (name string, meta int16) {
	if p.Cooked {
		return "minecraft:cooked_porkchop", 0
	}
	return "minecraft:porkchop", 0
}
