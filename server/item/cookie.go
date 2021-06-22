package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Cookie is a food item that can be obtained in large quantities, but do not restore hunger or saturation significantly.
type Cookie struct{}

// AlwaysConsumable ...
func (Cookie) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (Cookie) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (Cookie) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 0.4)
	return Stack{}
}

// EncodeItem ...
func (Cookie) EncodeItem() (name string, meta int16) {
	return "minecraft:cookie", 0
}
