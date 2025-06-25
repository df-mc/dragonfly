package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Wither is a lasting effect that causes an entity to take continuous damage
// that is capable of killing an entity.
var Wither wither

type wither struct {
	nopLasting
}

// Apply ...
func (wither) Apply(e world.Entity, eff Effect) {
	interval := max(80>>eff.Level(), 1)
	if eff.Tick()%interval == 0 {
		if l, ok := e.(living); ok {
			l.Hurt(1, WitherDamageSource{})
		}
	}
}

// RGBA ...
func (wither) RGBA() color.RGBA {
	return color.RGBA{R: 0x73, G: 0x61, B: 0x56, A: 0xff}
}

// WitherDamageSource is used for damage caused by an effect.wither applied
// to an entity.
type WitherDamageSource struct{}

func (WitherDamageSource) ReducedByResistance() bool { return true }
func (WitherDamageSource) ReducedByArmour() bool     { return false }
func (WitherDamageSource) Fire() bool                { return false }
func (WitherDamageSource) IgnoreTotem() bool         { return false }
