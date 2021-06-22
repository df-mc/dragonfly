package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// GoldenCarrot is a valuable food item and brewing ingredient.
type GoldenCarrot struct{}

// AlwaysConsumable ...
func (GoldenCarrot) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (GoldenCarrot) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (GoldenCarrot) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 14.4)
	return Stack{}
}

// EncodeItem ...
func (GoldenCarrot) EncodeItem() (name string, meta int16) {
	return "minecraft:golden_carrot", 0
}
