package item

import "github.com/df-mc/dragonfly/server/world"

// CookedChicken is a food item that can be eaten by the player.
type CookedChicken struct {
	defaultFood
}

// Consume ...
func (CookedChicken) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 7.2)
	return Stack{}
}

// EncodeItem ...
func (CookedChicken) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_chicken", 0
}
