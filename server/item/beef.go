package item

import "github.com/df-mc/dragonfly/server/world"

// Beef is a food item obtained from cows. It can be cooked in a furnace, smoker, or campfire.
type Beef struct {
	defaultFood

	// Cooked is whether the beef is cooked.
	Cooked bool
}

func (b Beef) Consume(_ *world.Tx, c Consumer) Stack {
	if b.Cooked {
		c.Saturate(8, 12.8)
	} else {
		c.Saturate(3, 1.8)
	}
	return Stack{}
}

func (b Beef) SmeltInfo() SmeltInfo {
	if b.Cooked {
		return SmeltInfo{}
	}
	return newFoodSmeltInfo(NewStack(Beef{Cooked: true}, 1), 0.35)
}

func (b Beef) EncodeItem() (name string, meta int16) {
	if b.Cooked {
		return "minecraft:cooked_beef", 0
	}
	return "minecraft:beef", 0
}
