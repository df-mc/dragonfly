package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// CookedSalmon is a food item obtained by cooking raw salmon.
type CookedSalmon struct{}

// AlwaysConsumable ...
func (CookedSalmon) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (CookedSalmon) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (CookedSalmon) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(6, 9.6)
	return Stack{}
}

// EncodeItem ...
func (CookedSalmon) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_salmon", 0
}
