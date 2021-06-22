package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// RawSalmon is a food item.
type RawSalmon struct{}

// AlwaysConsumable ...
func (RawSalmon) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RawSalmon) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RawSalmon) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 0.4)
	return Stack{}
}

// EncodeItem ...
func (RawSalmon) EncodeItem() (name string, meta int16) {
	return "minecraft:salmon", 0
}
