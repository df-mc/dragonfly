package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// WaterBreathing is a lasting effect that allows the affected entity to breath underwater until the effect
// expires.
type WaterBreathing struct {
	lastingEffect
}

// WithDurationAndLevel ...
func (w WaterBreathing) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return WaterBreathing{w.withDurationAndLevel(d, level)}
}

// RGBA ...
func (w WaterBreathing) RGBA() color.RGBA {
	return color.RGBA{R: 0x2e, G: 0x52, B: 0x99, A: 0xff}
}
