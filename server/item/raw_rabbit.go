package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// RawRabbit is a food item that can be eaten by the player, or cooked in a furnace or a campfire to make cooked rabbit.
type RawRabbit struct{}

// AlwaysConsumable ...
func (RawRabbit) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RawRabbit) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RawRabbit) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(3, 1.8)
	return Stack{}
}

// EncodeItem ...
func (RawRabbit) EncodeItem() (name string, meta int16) {
	return "minecraft:rabbit", 0
}
