package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// DriedKelp is a food item that can be quickly eaten by the player.
type DriedKelp struct{}

// AlwaysConsumable ...
func (DriedKelp) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (DriedKelp) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration / 2
}

// Consume ...
func (DriedKelp) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(1, 0.2)
	return Stack{}
}

// EncodeItem ...
func (DriedKelp) EncodeItem() (name string, meta int16) {
	return "minecraft:dried_kelp", 0
}
