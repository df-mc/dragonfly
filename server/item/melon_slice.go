package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// MelonSlice is a food item dropped by melon blocks.
type MelonSlice struct{}

func (m MelonSlice) AlwaysConsumable() bool {
	return false
}

func (m MelonSlice) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

func (m MelonSlice) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(2, 1.2)
	return Stack{}
}

func (MelonSlice) CompostChance() float64 {
	return 0.5
}

func (m MelonSlice) EncodeItem() (name string, meta int16) {
	return "minecraft:melon_slice", 0
}
