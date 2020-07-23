package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/healing"
	"image/color"
	"time"
)

// Regeneration is an effect that causes the entity that it is added to to slowly regenerate health. The level
// of the effect influences the speed with which the entity regenerates.
type Regeneration struct {
	lastingEffect
}

// Apply applies health to the entity.Living passed if the duration of the effect is at the right tick.
func (r Regeneration) Apply(e entity.Living) {
	interval := 50 >> r.Lvl
	if tickDuration(r.Dur)%interval == 0 {
		e.Heal(1, healing.SourceRegenerationEffect{})
	}
}

// WithDurationAndLevel ...
func (r Regeneration) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return Regeneration{r.withDurationAndLevel(d, level)}
}

// RGBA ...
func (Regeneration) RGBA() color.RGBA {
	return color.RGBA{R: 0xcd, G: 0x5c, B: 0xab, A: 0xff}
}
