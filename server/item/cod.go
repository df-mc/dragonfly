package item

import "github.com/df-mc/dragonfly/server/world"

// Cod is a food item obtained from cod. It can be cooked in a furnace, smoker, or campfire.
type Cod struct {
	defaultFood

	// Cooked is whether the cod is cooked.
	Cooked bool
}

func (c Cod) Consume(_ *world.Tx, co Consumer) Stack {
	if c.Cooked {
		co.Saturate(5, 6)
	} else {
		co.Saturate(2, 0.4)
	}
	return Stack{}
}

func (c Cod) SmeltInfo() SmeltInfo {
	if c.Cooked {
		return SmeltInfo{}
	}
	return newFoodSmeltInfo(NewStack(Cod{Cooked: true}, 1), 0.35)
}

func (c Cod) EncodeItem() (name string, meta int16) {
	if c.Cooked {
		return "minecraft:cooked_cod", 0
	}
	return "minecraft:cod", 0
}
