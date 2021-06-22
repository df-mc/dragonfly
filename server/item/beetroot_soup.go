package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Beetroot soup is an unstackable food item.
type BeetrootSoup struct{}

// MaxCount ...
func (BeetrootSoup) MaxCount() int {
	return 1
}

// AlwaysConsumable ...
func (BeetrootSoup) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (BeetrootSoup) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (BeetrootSoup) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 7.2)
	return NewStack(Bowl{}, 1)
}

// EncodeItem ...
func (BeetrootSoup) EncodeItem() (name string, meta int16) {
	return "minecraft:beetroot_soup", 0
}
