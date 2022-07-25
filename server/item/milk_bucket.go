package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// MilkBucket is an item that clears effects on consumption.
type MilkBucket struct{}

// MaxCount ...
func (m MilkBucket) MaxCount() int {
	return 1
}

// AlwaysConsumable ...
func (m MilkBucket) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (m MilkBucket) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (m MilkBucket) Consume(_ *world.World, c Consumer) Stack {
	for _, effect := range c.Effects() {
		c.RemoveEffect(effect.Type())
	}
	return NewStack(Bucket{}, 1)
}

// EncodeItem ...
func (m MilkBucket) EncodeItem() (name string, meta int16) {
	return "minecraft:milk_bucket", 0
}
