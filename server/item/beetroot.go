package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Beetroot is a food and dye ingredient.
type Beetroot struct{}

// AlwaysConsumable ...
func (b Beetroot) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (b Beetroot) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (b Beetroot) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(1, 1.2)
	return Stack{}
}

// EncodeItem ...
func (b Beetroot) EncodeItem() (name string, meta int16) {
	return "minecraft:beetroot", 0
}
