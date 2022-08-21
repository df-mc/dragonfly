package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Beetroot is a food and dye ingredient.
type Beetroot struct {
	defaultFood
}

// Consume ...
func (b Beetroot) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(1, 1.2)
	return Stack{}
}

// CompostChance ...
func (Beetroot) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (b Beetroot) EncodeItem() (name string, meta int16) {
	return "minecraft:beetroot", 0
}
