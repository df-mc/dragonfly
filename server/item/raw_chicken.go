package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
	"time"
)

// RawChicken is a food item that can be eaten by the player. It can be cooked in a furnace, smoker, or a campfire to make cooked chicken.
type RawChicken struct{}

// AlwaysConsumable ...
func (RawChicken) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (RawChicken) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (RawChicken) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 1.2)
	if rand.Float64() < 0.3 {
		c.AddEffect(effect.Hunger{}.WithSettings(30*time.Second, 1, false))
	}
	return Stack{}
}

// EncodeItem ...
func (RawChicken) EncodeItem() (name string, meta int16) {
	return "minecraft:chicken", 0
}
