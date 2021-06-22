package item

import (
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// CookedCod is a food item obtained by cooking raw cod.
type CookedCod struct{}

// AlwaysConsumable ...
func (CookedCod) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (CookedCod) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
}

// Consume ...
func (CookedCod) Consume(_ *world.World, c Consumer) Stack {
	c.Saturate(5, 6)
	return Stack{}
}

// EncodeItem ...
func (CookedCod) EncodeItem() (name string, meta int16) {
	return "minecraft:cooked_cod", 0
}
