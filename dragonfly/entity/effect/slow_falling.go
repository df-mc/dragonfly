package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// SlowFalling is a lasting effect that causes the affected entity to fall very slowly.
type SlowFalling struct {
	lastingEffect
}

// WithDurationAndLevel ...
func (s SlowFalling) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return SlowFalling{s.withDurationAndLevel(d, level)}
}

// RGBA ...
func (SlowFalling) RGBA() color.RGBA {
	return color.RGBA{R: 0xf7, G: 0xf8, B: 0xe0, A: 0xff}
}
