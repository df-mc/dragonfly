package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Pufferfish is a poisonous type of fish that is used to brew water breathing potions.
type Pufferfish struct{}

// AlwaysConsumable ...
func (p Pufferfish) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (p Pufferfish) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (p Pufferfish) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(1, 0.2)
	c.AddEffect(effect.Hunger{}.WithSettings(time.Duration(15)*time.Second, 3, false))
	c.AddEffect(effect.Poison{}.WithSettings(time.Duration(1)*time.Minute, 4, false))
	c.AddEffect(effect.Nausea{}.WithSettings(time.Duration(15)*time.Second, 1, false))
	return Stack{}
}

// EncodeItem ...
func (p Pufferfish) EncodeItem() (name string, meta int16) {
	return "minecraft:pufferfish", 0
}
