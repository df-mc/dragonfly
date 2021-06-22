package item

import "github.com/df-mc/dragonfly/server/world"

// RawChicken is a food item that can be eaten by the player. It can be cooked in a furnace, smoker, or a campfire to make cooked chicken.
type RawChicken struct {
	defaultFood
}

// Consume ...
func (RawChicken) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 1.2)
	return Stack{}
}

// EncodeItem ...
func (RawChicken) EncodeItem() (name string, meta int16) {
	return "minecraft:chicken", 0
}
