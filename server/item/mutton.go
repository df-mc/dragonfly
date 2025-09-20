package item

import "github.com/df-mc/dragonfly/server/world"

// Mutton is a food item obtained from sheep. It can be cooked in a furnace, smoker, or campfire.
type Mutton struct {
	defaultFood

	// Cooked is whether the mutton is cooked.
	Cooked bool
}

func (m Mutton) Consume(_ *world.Tx, c Consumer) Stack {
	if m.Cooked {
		c.Saturate(6, 9.6)
	} else {
		c.Saturate(2, 1.2)
	}
	return Stack{}
}

func (m Mutton) SmeltInfo() SmeltInfo {
	if m.Cooked {
		return SmeltInfo{}
	}
	return newFoodSmeltInfo(NewStack(Mutton{Cooked: true}, 1), 0.35)
}

func (m Mutton) EncodeItem() (name string, meta int16) {
	if m.Cooked {
		return "minecraft:cooked_mutton", 0
	}
	return "minecraft:mutton", 0
}
