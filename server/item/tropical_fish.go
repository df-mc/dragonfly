package item

import "github.com/df-mc/dragonfly/server/world"

// TropicalFish is a food item that cannot be cooked.
type TropicalFish struct {
	defaultFood
}

func (TropicalFish) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(1, 0.2)
	return Stack{}
}

func (TropicalFish) EncodeItem() (name string, meta int16) {
	return "minecraft:tropical_fish", 0
}
