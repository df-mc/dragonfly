package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// PumpkinPie is a food item that can be eaten by the player.
type PumpkinPie struct{}

// AlwaysConsumable ...
func (PumpkinPie) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (PumpkinPie) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (PumpkinPie) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(8, 4.8)
	return Stack{}
}

// EncodeItem ...
func (PumpkinPie) EncodeItem() (name string, meta int16) {
	return "minecraft:pumpkin_pie", 0
}
