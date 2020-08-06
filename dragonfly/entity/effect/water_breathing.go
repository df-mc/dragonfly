package effect

import (
	"image/color"
	"time"
)

// WaterBreathing is a lasting effect that allows the affected entity to breath underwater until the effect
// expires.
type WaterBreathing struct {
	lastingEffect
}

// WithSettings ...
func (w WaterBreathing) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return WaterBreathing{w.withSettings(d, level, ambient)}
}

// RGBA ...
func (w WaterBreathing) RGBA() color.RGBA {
	return color.RGBA{R: 0x2e, G: 0x52, B: 0x99, A: 0xff}
}
