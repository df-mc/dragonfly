package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// BakedPotato is a food item that can be eaten by the player.
type BakedPotato struct{}

// AlwaysConsumable ...
func (BakedPotato) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (BakedPotato) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (BakedPotato) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

// EncodeItem ...
func (BakedPotato) EncodeItem() (name string, meta int16) {
	return "minecraft:baked_potato", 0
}
