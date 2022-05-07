package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// GoldenApple is a special food item that bestows beneficial effects.
type GoldenApple struct{}

// AlwaysConsumable ...
func (e GoldenApple) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (e GoldenApple) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (e GoldenApple) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 9.6)
	c.AddEffect(effect.New(effect.Absorption{}, 1, 2*time.Minute))
	c.AddEffect(effect.New(effect.Regeneration{}, 2, 5*time.Second))
	return Stack{}
}

// EncodeItem ...
func (e GoldenApple) EncodeItem() (name string, meta int16) {
	return "minecraft:golden_apple", 0
}
