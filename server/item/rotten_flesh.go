package item

import "github.com/df-mc/dragonfly/server/world"

// RottenFlesh is a food item that can be eaten by the player, at the high risk of inflicting Hunger.
type RottenFlesh struct {
	defaultFood
}

// Consume ...
func (RottenFlesh) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 0.8)
	return Stack{}
}

// EncodeItem ...
func (RottenFlesh) EncodeItem() (name string, meta int16) {
	return "minecraft:rotten_flesh", 0
}
