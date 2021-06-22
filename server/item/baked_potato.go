package item

import "github.com/df-mc/dragonfly/server/world"

// BakedPotato is a food item that can be eaten by the player.
type BakedPotato struct {
	defaultFood
}

// Consume ...
func (BakedPotato) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

// EncodeItem ...
func (BakedPotato) EncodeItem() (name string, meta int16) {
	return "minecraft:baked_potato", 0
}
