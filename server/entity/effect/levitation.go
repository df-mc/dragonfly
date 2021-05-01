package effect

import (
	"image/color"
	"time"
)

// Levitation is a lasting effect that causes the affected entity to slowly levitate upwards. It is roughly
// the opposite of the SlowFalling effect.
type Levitation struct {
	lastingEffect
}

// WithSettings ...
func (l Levitation) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Levitation{l.withSettings(d, level, ambient)}
}

// RGBA ...
func (Levitation) RGBA() color.RGBA {
	return color.RGBA{R: 0xce, G: 0xff, B: 0xff, A: 0xff}
}
