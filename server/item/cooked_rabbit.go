package item

import "github.com/df-mc/dragonfly/server/world"

// CookedRabbit is a food item that can be eaten by the player.
type CookedRabbit struct {
	defaultFood
}

// Consume ...
func (CookedRabbit) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

// EncodeItem ...
func (CookedRabbit) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_rabbit", 0
}
