package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Beetroot is a food and dye ingredient.
type Beetroot struct {
	defaultFood
}

func (b Beetroot) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(1, 1.2)
	return Stack{}
}

func (Beetroot) CompostChance() float64 {
	return 0.65
}

func (b Beetroot) EncodeItem() (name string, meta int16) {
	return "minecraft:beetroot", 0
}
