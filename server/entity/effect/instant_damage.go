package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// InstantDamage is an instant effect that causes a living entity to
// immediately take some damage, depending on the level and the potency of the
// effect.
var InstantDamage instantDamage

type instantDamage struct{}

// Apply ...
func (i instantDamage) Apply(e world.Entity, eff Effect) {
	base := 3 << eff.Level()
	if l, ok := e.(living); ok {
		l.Hurt(float64(base)*eff.potency, InstantDamageSource{})
	}
}

// RGBA ...
func (instantDamage) RGBA() color.RGBA {
	return color.RGBA{R: 0xa9, G: 0x65, B: 0x6a, A: 0xff}
}

// InstantDamageSource is used for damage caused by an effect.instantDamage
// applied to an entity.
type InstantDamageSource struct{}

func (InstantDamageSource) ReducedByArmour() bool     { return false }
func (InstantDamageSource) ReducedByResistance() bool { return true }
func (InstantDamageSource) Fire() bool                { return false }
func (InstantDamageSource) IgnoreTotem() bool         { return false }
