package item

import "github.com/df-mc/dragonfly/server/world"

// Bread is a food item that can be eaten by the player.
type Bread struct {
	defaultFood
}

// Consume ...
func (Bread) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

// CompostChance ...
func (Bread) CompostChance() float64 {
	return 0.85
}

// EncodeItem ...
func (Bread) EncodeItem() (name string, meta int16) {
	return "minecraft:bread", 0
}
