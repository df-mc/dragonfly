package item

import "github.com/df-mc/dragonfly/server/world"

// CookedPorkchop is a food item that can be eaten by the player.
type CookedPorkchop struct {
	defaultFood
}

// Consume ...
func (CookedPorkchop) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(8, 12.8)
	return Stack{}
}

// EncodeItem ...
func (CookedPorkchop) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_porkchop", 0
}
