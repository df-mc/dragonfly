package item

import (
	"github.com/df-mc/dragonfly/server/item/potion"
	"github.com/df-mc/dragonfly/server/world"
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
	for _, effect := range p.Type.Effects() {
		c.AddEffect(effect)
	}
	return NewStack(GlassBottle{}, 1)
}

// EncodeItem ...
func (p Potion) EncodeItem() (name string, meta int16) {
	return "minecraft:potion", int16(p.Type.Uint8())
}
