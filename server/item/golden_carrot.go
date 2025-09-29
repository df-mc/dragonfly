package item

import "github.com/df-mc/dragonfly/server/world"

// GoldenCarrot is a valuable food item and brewing ingredient. It provides the second most saturation in the game,
// behind Suspicious Stew crafted with either a Dandelion or Blue Orchid.
type GoldenCarrot struct {
	defaultFood
}

func (GoldenCarrot) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(6, 14.4)
	return Stack{}
}

func (GoldenCarrot) EncodeItem() (name string, meta int16) {
	return "minecraft:golden_carrot", 0
}
