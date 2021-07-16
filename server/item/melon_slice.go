package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// MelonSlice is a food item dropped by melon blocks.
type MelonSlice struct{}

// AlwaysConsumable ...
func (m MelonSlice) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (m MelonSlice) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (m MelonSlice) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 1.2)
	return Stack{}
}

// EncodeItem ...
func (m MelonSlice) EncodeItem() (name string, meta int16) {
	return "minecraft:melon_slice", 0
}
