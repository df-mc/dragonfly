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

// WithDuration ...
func (l Levitation) WithDuration(d time.Duration) entity.Effect {
	return Levitation{l.withDuration(d)}
}

// RGBA ...
func (Levitation) RGBA() color.RGBA {
	return color.RGBA{R: 0xce, G: 0xff, B: 0xff, A: 0xff}
}
