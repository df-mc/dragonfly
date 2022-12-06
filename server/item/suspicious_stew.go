package item

import (
	"github.com/df-mc/dragonfly/server/world"
)

// SuspiciousStew is a food item that can give the player a status effect that depends on the flower used to craft it.
type SuspiciousStew struct {
	defaultFood

	// Type specifies the type of effect will be given to the player
	Type StewType
}

// MaxCount ...
func (SuspiciousStew) MaxCount() int {
	return 1
}

// AlwaysConsumable ...
func (SuspiciousStew) AlwaysConsumable() bool {
	return true
}

// EncodeItem ...
func (s SuspiciousStew) EncodeItem() (name string, meta int16) {
	return "minecraft:suspicious_stew", int16(s.Type.Uint8())
}

// Consume ...
func (s SuspiciousStew) Consume(_ *world.World, c Consumer) Stack {
	for _, effect := range s.Type.Effects() {
		c.AddEffect(effect)
	}
	c.Saturate(6, 7.2)

	return NewStack(Bowl{}, 1)
}
