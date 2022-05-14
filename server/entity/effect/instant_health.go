package effect

import (
	"github.com/df-mc/dragonfly/server/entity/healing"
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// InstantHealth is an instant effect that causes the player that it is applied to immediately regain some
// health. The amount of health regained depends on the effect level and potency.
type InstantHealth struct {
	// Potency specifies the potency of the instant health. By default, this value is 1, which means 100% of
	// the instant health will be applied to an entity. A lingering health potion, for example, has a potency
	// of 0.5: It heals 1 heart (per tick) instead of 2.
	Potency float64
}

// WithPotency ...
func (i InstantHealth) WithPotency(potency float64) Type {
	i.Potency = potency
	return i
}

// Apply instantly heals the world.Entity passed for a bit of health, depending on the effect level and
// potency.
func (i InstantHealth) Apply(e world.Entity, lvl int, _ time.Duration) {
	if i.Potency == 0 {
		// Potency of 1 by default.
		i.Potency = 1
	}
	base := 2 << lvl
	if l, ok := e.(living); ok {
		l.Heal(float64(base)*i.Potency, healing.SourceInstantHealthEffect{})
	}
}

// RGBA ...
func (InstantHealth) RGBA() color.RGBA {
	return color.RGBA{R: 0xf8, G: 0x24, B: 0x23, A: 0xff}
}
