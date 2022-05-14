package effect

import (
	"github.com/df-mc/dragonfly/server/entity/damage"
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// InstantDamage is an instant effect that causes a living entity to immediately take some damage, depending
// on the level and the potency of the effect.
type InstantDamage struct {
	// Potency specifies the potency of the instant damage. By default, this value is 1, which means 100% of
	// the instant damage will be applied to an entity. A lingering damage potion, for example, has a potency
	// of 0.5: It deals 1.5 hearts damage (per tick) instead of 3.
	Potency float64
}

// WithPotency ...
func (i InstantDamage) WithPotency(potency float64) Type {
	i.Potency = potency
	return i
}

// Apply ...
func (i InstantDamage) Apply(e world.Entity, lvl int, _ time.Duration) {
	if i.Potency == 0 {
		// Potency of 1 by default.
		i.Potency = 1
	}
	base := 3 << lvl
	if l, ok := e.(living); ok {
		l.Hurt(float64(base)*i.Potency, damage.SourceInstantDamageEffect{})
	}
}

// RGBA ...
func (InstantDamage) RGBA() color.RGBA {
	return color.RGBA{R: 0x43, G: 0x0a, B: 0x09, A: 0xff}
}
