package item

import "github.com/df-mc/dragonfly/server/world"

// BakedPotato is a food item that can be eaten by the player.
type BakedPotato struct {
	defaultFood
}

func (BakedPotato) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

func (BakedPotato) CompostChance() float64 {
	return 0.85
}

func (BakedPotato) EncodeItem() (name string, meta int16) {
	return "minecraft:baked_potato", 0
}
