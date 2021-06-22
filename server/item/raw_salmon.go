package item

import "github.com/df-mc/dragonfly/server/world"

// RawSalmon is a food item.
type RawSalmon struct {
	defaultFood
}

// Consume ...
func (RawSalmon) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 0.4)
	return Stack{}
}

// EncodeItem ...
func (RawSalmon) EncodeItem() (name string, meta int16) {
	return "minecraft:salmon", 0
}
