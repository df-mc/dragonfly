package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"image/color"
	"time"
)

// Wither is a lasting effect that causes an entity to take continuous damage that is capable of killing an
// entity.
type Wither struct {
	lastingEffect
}

// Apply ...
func (w Wither) Apply(e entity.Living) {
	interval := 80 >> w.Lvl
	if tickDuration(w.Dur)%interval == 0 {
		e.Hurt(1, damage.SourceWitherEffect{})
	}
}

// WithDuration ...
func (w Wither) WithDuration(d time.Duration) entity.Effect {
	return Wither{w.withDuration(d)}
}

// RGBA ...
func (Wither) RGBA() color.RGBA {
	return color.RGBA{R: 0x35, G: 0x2a, B: 0x27, A: 0xff}
}
