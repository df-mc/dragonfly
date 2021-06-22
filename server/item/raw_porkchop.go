package item

import "github.com/df-mc/dragonfly/server/world"

// RawPorkchop is a food item that can be eaten by the player or cooked to make a cooked porkchop.
type RawPorkchop struct {
	defaultFood
}

// Consume ...
func (RawPorkchop) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(3, 1.8)
	return Stack{}
}

// EncodeItem ...
func (RawPorkchop) EncodeItem() (name string, meta int16) {
	return "minecraft:porkchop", 0
}
