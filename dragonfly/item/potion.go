package item

import (
	"github.com/df-mc/dragonfly/dragonfly/item/potion"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// Potion is a consumable item that grant effects.
type Potion struct {
	// Type is the type of potion.
	Type potion.Potion
}

// AlwaysConsumable ...
func (p Potion) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (p Potion) ConsumeDuration() time.Duration {
	return defaultConsumeDuration
}

// Consume ...
func (p Potion) Consume(_ *world.World, c Consumer) Stack {
	for _, effect := range p.Type.Effects {
		c.AddEffect(effect)
	}
	return NewStack(GlassBottle{}, 1)
}

// EncodeItem ...
func (p Potion) EncodeItem() (id int32, meta int16) {
	return 373, int16(p.Type.Uint8())
}
