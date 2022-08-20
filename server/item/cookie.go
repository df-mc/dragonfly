package item

import "github.com/df-mc/dragonfly/server/world"

// Cookie is a food item that can be obtained in large quantities, but do not restore hunger or saturation significantly.
type Cookie struct {
	defaultFood
}

// Consume ...
func (Cookie) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 0.4)
	return Stack{}
}

// CompostChance ...
func (Cookie) CompostChance() float64 {
	return 0.85
}

// EncodeItem ...
func (Cookie) EncodeItem() (name string, meta int16) {
	return "minecraft:cookie", 0
}
