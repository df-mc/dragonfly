package item

import "github.com/df-mc/dragonfly/server/world"

// RawMutton is a food item dropped by sheep when killed.
type RawMutton struct {
	defaultFood
}

// Consume ...
func (RawMutton) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 1.2)
	return Stack{}
}

// EncodeItem ...
func (RawMutton) EncodeItem() (name string, meta int16) {
	return "minecraft:mutton", 0
}
