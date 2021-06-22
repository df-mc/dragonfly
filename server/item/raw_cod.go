package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// RawCod is a food item.
type RawCod struct{}

// AlwaysConsumable ...
func (RawCod) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RawCod) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RawCod) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 0.4)
	return Stack{}
}

// EncodeItem ...
func (RawCod) EncodeItem() (name string, meta int16) {
	return "minecraft:cod", 0
}
