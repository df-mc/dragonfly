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
	c.AddEffect(effect.Absorption{}.WithSettings(2*time.Minute, 4, false))
	c.AddEffect(effect.Regeneration{}.WithSettings(30*time.Second, 5, false))
	c.AddEffect(effect.FireResistance{}.WithSettings(5*time.Minute, 1, false))
	c.AddEffect(effect.Resistance{}.WithSettings(5*time.Minute, 1, false))
	return Stack{}
}

// EncodeItem ...
func (EnchantedApple) EncodeItem() (name string, meta int16) {
	return "minecraft:enchanted_golden_apple", 0
}
