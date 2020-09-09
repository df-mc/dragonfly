package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"image/color"
	"time"
)

// Poison is a lasting effect that causes the affected entity to lose health gradually. Poison cannot kill,
// unlike FatalPoison.
type Poison struct {
	lastingEffect
}

// Apply ...
func (p Poison) Apply(e world.Entity) {
	interval := 50 >> p.Lvl
	if tickDuration(p.Dur)%interval == 0 {
		if living, ok := e.(living); ok {
			living.Hurt(1, damage.SourcePoisonEffect{})
		}
	}
}

// WithSettings ...
func (p Poison) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Poison{p.withSettings(d, level, ambient)}
}

// RGBA ...
func (p Poison) RGBA() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x93, B: 0x31, A: 0xff}
}
