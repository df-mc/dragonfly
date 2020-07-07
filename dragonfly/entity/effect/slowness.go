package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Slowness is a lasting effect that decreases the movement speed of a living entity by 15% for each level
// that the effect has.
type Slowness struct {
	lastingEffect
}

// Start ...
func (s Slowness) Start(e entity.Living) {
	slowness := 1 - float64(s.Lvl)*0.15
	e.SetSpeed(e.Speed() * slowness)
}

// Stop ...
func (s Slowness) Stop(e entity.Living) {
	slowness := 1 - float64(s.Lvl)*0.15
	e.SetSpeed(e.Speed() / slowness)
}

// WithDuration ...
func (s Slowness) WithDuration(d time.Duration) entity.Effect {
	return Slowness{s.withDuration(d)}
}

// RGBA ...
func (s Slowness) RGBA() color.RGBA {
	return color.RGBA{R: 0x5a, G: 0x6c, B: 0x81, A: 0xff}
}
