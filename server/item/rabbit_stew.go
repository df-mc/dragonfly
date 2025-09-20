package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// RabbitStew is a food item that can be eaten by the player.
type RabbitStew struct {
	defaultFood
}

func (RabbitStew) MaxCount() int {
	return 1
}

func (RabbitStew) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(10, 12)
	return NewStack(Bowl{}, 1)
}

func (RabbitStew) EncodeItem() (name string, meta int16) {
	return "minecraft:rabbit_stew", 0
}
