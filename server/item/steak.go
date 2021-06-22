package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// Steak is a food item obtained from cows or from cooking raw beef.
type Steak struct{}

// AlwaysConsumable ...
func (Steak) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (Steak) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (Steak) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(8, 12.8)
	return Stack{}
}

// EncodeItem ...
func (Steak) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_beef", 0
}
