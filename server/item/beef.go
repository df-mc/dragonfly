package item

import "github.com/df-mc/dragonfly/server/world"

// Beef is a food item obtained from cows. It can be cooked in a furnace, smoker, or campfire.
type Beef struct {
	defaultFood

	// Cooked is whether the beef is cooked.
	Cooked bool
}

// Consume ...
func (b Beef) Consume(_ *world.World, c Consumer) Stack {
	if b.Cooked {
		c.Saturate(8, 12.8)
	} else {
		c.Saturate(3, 1.8)
	}
	return Stack{}
}

// EncodeItem ...
func (b Beef) EncodeItem() (name string, meta int16) {
	if b.Cooked {
		return "minecraft:cooked_beef", 0
	}
	return "minecraft:beef", 0
}
