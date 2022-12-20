package effect

import (
	"github.com/df-mc/dragonfly/server/world"
	"image/color"
	"time"
)

// Poison is a lasting effect that causes the affected entity to lose health gradually. Poison cannot kill,
// unlike FatalPoison.
type Poison struct {
	nopLasting
}

// Apply ...
func (Poison) Apply(e world.Entity, lvl int, d time.Duration) {
	interval := 50 >> (lvl - 1)
	if interval < 1 {
		interval = 1
	}
	if tickDuration(d)%interval == 0 {
		if l, ok := e.(living); ok && l.Health() > 1 {
			l.Hurt(1, PoisonDamageSource{})
		}
	}
}

// RGBA ...
func (Poison) RGBA() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x93, B: 0x31, A: 0xff}
}

// PoisonDamageSource is used for damage caused by an effect.Poison or
// effect.FatalPoison applied to an entity.
type PoisonDamageSource struct {
	// Fatal specifies if the damage was caused by effect.FatalPoison and if
	// the damage could therefore kill the entity.
	Fatal bool
}

func (PoisonDamageSource) ReducedByResistance() bool { return true }
func (PoisonDamageSource) ReducedByArmour() bool     { return false }
func (PoisonDamageSource) Fire() bool                { return false }
