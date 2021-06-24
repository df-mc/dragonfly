package item

import (
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// SpiderEye is a poisonous food and brewing item.
type SpiderEye struct{}

// AlwaysConsumable ...
func (SpiderEye) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (SpiderEye) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (SpiderEye) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(2, 3.2)
	c.AddEffect(effect.Poison{}.WithSettings(time.Second*5, 1, false))
	return Stack{}
}

// EncodeItem ...
func (SpiderEye) EncodeItem() (name string, meta int16) {
	return "minecraft:spider_eye", 0
}
