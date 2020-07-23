package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Levitation is a lasting effect that causes the affected entity to slowly levitate upwards. It is roughly
// the opposite of the SlowFalling effect.
type Levitation struct {
	lastingEffect
}

// WithDurationAndLevel ...
func (l Levitation) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return Levitation{l.withDurationAndLevel(d, level)}
}

// RGBA ...
func (Levitation) RGBA() color.RGBA {
	return color.RGBA{R: 0xce, G: 0xff, B: 0xff, A: 0xff}
}
