package item

import "github.com/df-mc/dragonfly/server/world"

// Rabbit is a food item obtained from rabbits. It can be cooked in a furnace, smoker, or campfire.
type Rabbit struct {
	defaultFood

	// Cooked is whether the rabbit is cooked.
	Cooked bool
}

func (r Rabbit) Consume(_ *world.Tx, c Consumer) Stack {
	if r.Cooked {
		c.Saturate(5, 6)
	} else {
		c.Saturate(3, 1.8)
	}
	return Stack{}
}

func (r Rabbit) SmeltInfo() SmeltInfo {
	if r.Cooked {
		return SmeltInfo{}
	}
	return newFoodSmeltInfo(NewStack(Rabbit{Cooked: true}, 1), 0.35)
}

func (r Rabbit) EncodeItem() (name string, meta int16) {
	if r.Cooked {
		return "minecraft:cooked_rabbit", 0
	}
	return "minecraft:rabbit", 0
}
