package item

import "github.com/df-mc/dragonfly/server/world"

// CookedCod is a food item obtained by cooking raw cod.
type CookedCod struct {
	defaultFood
}

// Consume ...
func (CookedCod) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

// EncodeItem ...
func (CookedCod) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_cod", 0
}
