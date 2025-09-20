package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// DriedKelp is a food item that can be quickly eaten by the player.
type DriedKelp struct{}

func (DriedKelp) AlwaysConsumable() bool {
	return false
}

func (DriedKelp) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration / 2
}

func (DriedKelp) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(1, 0.2)
	return Stack{}
}

func (DriedKelp) CompostChance() float64 {
	return 0.3
}

func (DriedKelp) EncodeItem() (name string, meta int16) {
	return "minecraft:dried_kelp", 0
}
