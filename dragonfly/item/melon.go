package item

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// Melon is a food item dropped by melon blocks.
type Melon struct{}

// AlwaysConsumable ...
func (m Melon) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (m Melon) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (m Melon) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 1.2)
	return Stack{}
}

// EncodeItem ...
func (m Melon) EncodeItem() (id int32, meta int16) {
	return 360, 0
}
