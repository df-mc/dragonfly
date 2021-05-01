package item

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// EnchantedApple is a rare variant of the golden apple that has stronger effects.
type EnchantedApple struct{}

// AlwaysConsumable ...
func (e EnchantedApple) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (e EnchantedApple) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (e EnchantedApple) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(4, 9.6)
	c.AddEffect(effect.Absorption{}.WithSettings(time.Duration(2)*time.Minute, 4, false))
	c.AddEffect(effect.Regeneration{}.WithSettings(time.Duration(30)*time.Second, 5, false))
	c.AddEffect(effect.FireResistance{}.WithSettings(time.Duration(5)*time.Minute, 1, false))
	c.AddEffect(effect.Resistance{}.WithSettings(time.Duration(5)*time.Minute, 1, false))
	return Stack{}
}

// EncodeItem ...
func (e EnchantedApple) EncodeItem() (id int32, name string, meta int16) {
	return 466, "minecraft:appleenchanted", 0
}
