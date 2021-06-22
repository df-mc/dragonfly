package item

import "github.com/df-mc/dragonfly/server/world"

// RawCod is a food item.
type RawCod struct {
	defaultFood
}

// Consume ...
func (RawCod) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 0.4)
	return Stack{}
}

// EncodeItem ...
func (RawCod) EncodeItem() (name string, meta int16) {
	return "minecraft:cod", 0
}
