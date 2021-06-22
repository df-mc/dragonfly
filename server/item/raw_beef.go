package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// RawBeef is a food item that can be eaten by the player or cooked in a furnace, smoker, or campfire to make steak.
type RawBeef struct{}

// AlwaysConsumable ...
func (RawBeef) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RawBeef) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RawBeef) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(3, 1.8)
	return Stack{}
}

// EncodeItem ...
func (RawBeef) EncodeItem() (name string, meta int16) {
	return "minecraft:beef", 0
}
