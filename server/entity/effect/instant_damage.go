package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// InstantDamage is an instant effect that causes a living entity to immediately take some damage, depending
// on the level and the potency of the effect.
type InstantDamage struct{}

// Apply ...
func (i InstantDamage) Apply(e world.Entity, eff Effect) {
	base := 3 << eff.Level()
	if l, ok := e.(living); ok {
		l.Hurt(float64(base)*eff.potency, InstantDamageSource{})
	}
}

// RGBA ...
func (InstantDamage) RGBA() color.RGBA {
	return color.RGBA{R: 0x43, G: 0x0a, B: 0x09, A: 0xff}
}

// InstantDamageSource is used for damage caused by an effect.InstantDamage
// applied to an entity.
type InstantDamageSource struct{}

func (InstantDamageSource) ReducedByArmour() bool     { return false }
func (InstantDamageSource) ReducedByResistance() bool { return true }
func (InstantDamageSource) Fire() bool                { return false }
