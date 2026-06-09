package item

import (
	"time"

	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/world"
)

// HoneyBottle is a food item obtained from full beehives or beenests using a glass bottle. Consuming it
// restores hunger, removes any active poison and returns an empty glass bottle.
type HoneyBottle struct{}

// MaxCount ...
func (HoneyBottle) MaxCount() int {
	return 16
}

// AlwaysConsumable ...
func (HoneyBottle) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (HoneyBottle) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration * 5 / 4
}

// Consume ...
func (HoneyBottle) Consume(_ *world.Tx, c Consumer) Stack {
	c.Saturate(6, 1.2)
	c.RemoveEffect(effect.Poison)
	return NewStack(GlassBottle{}, 1)
}

// EncodeItem ...
func (HoneyBottle) EncodeItem() (name string, meta int16) {
	return "minecraft:honey_bottle", 0
}
