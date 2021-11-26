package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
	"time"
)

// PoisonousPotato is a type of potato that can poison the player.
type PoisonousPotato struct {
	defaultFood
}

// Consume ...
func (p PoisonousPotato) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 1.2)
	if rand.Float64() < 0.6 {
		c.AddEffect(effect.New(effect.Poison{}, 1, 5*time.Second))
	}
	return Stack{}
}

// EncodeItem ...
func (p PoisonousPotato) EncodeItem() (name string, meta int16) {
	return "minecraft:poisonous_potato", 0
}
