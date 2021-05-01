package item

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/effect"
	"github.com/df-mc/dragonfly/dragonfly/world"
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
	c.AddEffect(effect.Absorption{}.WithSettings(time.Duration(2)*time.Minute, 1, false))
	c.AddEffect(effect.Regeneration{}.WithSettings(time.Duration(5)*time.Minute, 2, false))
	return Stack{}
}

// EncodeItem ...
func (e GoldenApple) EncodeItem() (id int32, name string, meta int16) {
	return 322, "minecraft:golden_apple", 0
}
