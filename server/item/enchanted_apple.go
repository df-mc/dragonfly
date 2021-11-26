package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// EnchantedApple is a rare variant of the golden apple that has stronger effects.
type EnchantedApple struct{}

// AlwaysConsumable ...
func (EnchantedApple) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (EnchantedApple) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (EnchantedApple) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 9.6)
	c.AddEffect(effect.New(effect.Absorption{}, 4, 2*time.Minute))
	c.AddEffect(effect.New(effect.Regeneration{}, 5, 30*time.Second))
	c.AddEffect(effect.New(effect.FireResistance{}, 1, 5*time.Minute))
	c.AddEffect(effect.New(effect.Resistance{}, 1, 5*time.Minute))
	return Stack{}
}

// EncodeItem ...
func (EnchantedApple) EncodeItem() (name string, meta int16) {
	return "minecraft:enchanted_golden_apple", 0
}
