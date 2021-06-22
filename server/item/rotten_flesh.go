package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
	"time"
)

// RottenFlesh is a food item that can be eaten by the player, at the high risk of inflicting Hunger.
type RottenFlesh struct{}

// AlwaysConsumable ...
func (RottenFlesh) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RottenFlesh) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RottenFlesh) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 0.8)
	if rand.Float64() < 0.8 {
		c.AddEffect(effect.Hunger{}.WithSettings(30*time.Second, 1, false))
	}
	return Stack{}
}

// EncodeItem ...
func (RottenFlesh) EncodeItem() (name string, meta int16) {
	return "minecraft:rotten_flesh", 0
}
