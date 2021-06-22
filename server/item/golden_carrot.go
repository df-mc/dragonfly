package item

import "github.com/df-mc/dragonfly/server/world"

// GoldenCarrot is a valuable food item and brewing ingredient. It provides the second most saturation in the game,
// behind Suspicious Stew crafted with either a Dandelion or Blue Orchid.
type GoldenCarrot struct {
	defaultFood
}

// Consume ...
func (GoldenCarrot) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 14.4)
	return Stack{}
}

// EncodeItem ...
func (GoldenCarrot) EncodeItem() (name string, meta int16) {
	return "minecraft:golden_carrot", 0
}
