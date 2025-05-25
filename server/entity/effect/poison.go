package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
)

// Poison is a lasting effect that causes the affected entity to lose health
// gradually. poison cannot kill, unlike fatalPoison.
var Poison poison

type poison struct {
	nopLasting
}

// Apply ...
func (poison) Apply(e world.Entity, eff Effect) {
	interval := max(50>>(eff.Level()-1), 1)
	if eff.Tick()%interval == 0 {
		if l, ok := e.(living); ok && l.Health() > 1 {
			l.Hurt(1, PoisonDamageSource{})
		}
	}
}

// RGBA ...
func (poison) RGBA() color.RGBA {
	return color.RGBA{R: 0x87, G: 0xa3, B: 0x63, A: 0xff}
}

// PoisonDamageSource is used for damage caused by an effect.poison or
// effect.fatalPoison applied to an entity.
type PoisonDamageSource struct {
	// Fatal specifies if the damage was caused by effect.fatalPoison and if
	// the damage could therefore kill the entity.
	Fatal bool
}

func (PoisonDamageSource) ReducedByResistance() bool { return true }
func (PoisonDamageSource) ReducedByArmour() bool     { return false }
func (PoisonDamageSource) Fire() bool                { return false }
func (PoisonDamageSource) IgnoreTotem() bool         { return false }
