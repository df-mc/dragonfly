package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// MushroomStew is a food item.
type MushroomStew struct {
	defaultFood
}

// MaxCount ...
func (MushroomStew) MaxCount() int {
	return 1
}

// Consume ...
func (MushroomStew) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 7.2)
	return NewStack(Bowl{}, 1)
}

// EncodeItem ...
func (MushroomStew) EncodeItem() (name string, meta int16) {
	return "minecraft:mushroom_stew", 0
}
