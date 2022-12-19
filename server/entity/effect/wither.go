package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// Wither is a lasting effect that causes an entity to take continuous damage that is capable of killing an
// entity.
type Wither struct {
	nopLasting
}

// Apply ...
func (Wither) Apply(e world.Entity, lvl int, d time.Duration) {
	interval := 50 >> (lvl - 1)
	if interval < 1 {
		interval = 1
	}
	if tickDuration(d)%interval == 0 {
		if l, ok := e.(living); ok {
			l.Hurt(1, WitherDamageSource{})
		}
	}
}

// RGBA ...
func (Wither) RGBA() color.RGBA {
	return color.RGBA{R: 0x35, G: 0x2a, B: 0x27, A: 0xff}
}

// WitherDamageSource is used for damage caused by an effect.Wither applied
// to an entity.
type WitherDamageSource struct{}

func (WitherDamageSource) ReducedByResistance() bool { return true }
func (WitherDamageSource) ReducedByArmour() bool     { return false }
func (WitherDamageSource) Fire() bool                { return false }
