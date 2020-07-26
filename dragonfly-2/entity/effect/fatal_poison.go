package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"image/color"
	"time"
)

// FatalPoison is a lasting effect that causes the affected entity to lose health gradually. FatalPoison,
// unlike Poison, can kill the entity it is applied to.
type FatalPoison struct {
	lastingEffect
}

// Apply ...
func (p FatalPoison) Apply(e entity.Living) {
	interval := 50 >> p.Lvl
	if tickDuration(p.Dur)%interval == 0 {
		e.Hurt(1, damage.SourcePoisonEffect{Fatal: true})
	}
}

// WithSettings ...
func (p FatalPoison) WithSettings(d time.Duration, level int, ambient bool) entity.Effect {
	return FatalPoison{p.withSettings(d, level, ambient)}
}

// RGBA ...
func (p FatalPoison) RGBA() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x93, B: 0x31, A: 0xff}
}
