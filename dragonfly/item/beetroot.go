package item

import (
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// Beetroot is a food & dye ingredient.
type Beetroot struct{}

// AlwaysConsumable ...
func (b Beetroot) AlwaysConsumable() bool {
	return false
}

// ConsumeDuration ...
func (b Beetroot) ConsumeDuration() time.Duration {
	return defaultConsumeDuration
}

// Consume ...
func (b Beetroot) Consume(w *world.World, c Consumer) Stack {
	c.Saturate(1, 1.2)
	return Stack{}
}

// EncodeItem ...
func (b Beetroot) EncodeItem() (id int32, meta int16) {
	return 457, 0
}
