package item

import "github.com/df-mc/dragonfly/server/world"

// CookedSalmon is a food item obtained by cooking raw salmon. It is a nutritious and easily obtainable early-game food source.
type CookedSalmon struct {
	defaultFood
}

// Consume ...
func (CookedSalmon) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 9.6)
	return Stack{}
}

// EncodeItem ...
func (CookedSalmon) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_salmon", 0
}
