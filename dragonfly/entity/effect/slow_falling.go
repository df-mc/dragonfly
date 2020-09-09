package effect

import (
	"image/color"
	"time"
)

// SlowFalling is a lasting effect that causes the affected entity to fall very slowly.
type SlowFalling struct {
	lastingEffect
}

// WithSettings ...
func (s SlowFalling) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return SlowFalling{s.withSettings(d, level, ambient)}
}

// RGBA ...
func (SlowFalling) RGBA() color.RGBA {
	return color.RGBA{R: 0xf7, G: 0xf8, B: 0xe0, A: 0xff}
}
