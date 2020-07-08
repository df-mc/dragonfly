package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"image/color"
	"time"
)

// Poison is a lasting effect that causes the affected entity to lose health gradually. Poison cannot kill,
// unlike FatalPoison.
type Poison struct {
	lastingEffect
}

// Apply ...
func (p Poison) Apply(e entity.Living) {
	interval := 50 >> p.Lvl
	if tickDuration(p.Dur)%interval == 0 {
		e.Hurt(1, damage.SourcePoisonEffect{})
	}
}

// WithDuration ...
func (p Poison) WithDuration(d time.Duration) entity.Effect {
	return Poison{p.withDuration(d)}
}

// RGBA ...
func (p Poison) RGBA() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x93, B: 0x31, A: 0xff}
}
