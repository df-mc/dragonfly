package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// BeetrootSoup is an unstackable food item.
type BeetrootSoup struct {
	defaultFood
}

// MaxCount ...
func (BeetrootSoup) MaxCount() int {
	return 1
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
