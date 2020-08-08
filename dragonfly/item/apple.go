package item

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// Apple is a food item that can be eaten by the player.
type Apple struct{}

// AlwaysConsumable ...
func (a Apple) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (a Apple) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (a Apple) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 2.4)
	return Stack{}
}

// EncodeItem ...
func (a Apple) EncodeItem() (id int32, meta int16) {
	return 260, 0
}
