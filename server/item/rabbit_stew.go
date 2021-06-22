package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// RabbitStew is a food item that can be eaten by the player.
type RabbitStew struct{}

// MaxCount ...
func (RabbitStew) MaxCount() int {
	return 1
}

// AlwaysConsumable ...
func (RabbitStew) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RabbitStew) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RabbitStew) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(10, 12)
	return NewStack(Bowl{}, 1)
}

// EncodeItem ...
func (RabbitStew) EncodeItem() (name string, meta int16) {
	return "minecraft:rabbit_stew", 0
}
