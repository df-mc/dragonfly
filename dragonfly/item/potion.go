package item

import (
	"github.com/df-mc/dragonfly/dragonfly/item/potion"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"time"
)

// Potion is an item that grants effects on consumption.
type Potion struct {
	// Type is the type of potion.
	Type potion.Potion
}

// MaxCount ...
func (p Potion) MaxCount() int {
	return 1
}

// AlwaysConsumable ...
func (p Potion) AlwaysConsumable() bool {
	return true
}

// ConsumeDuration ...
func (p Potion) ConsumeDuration() time.Duration {
	return DefaultConsumeDuration
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
