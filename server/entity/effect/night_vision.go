package effect

import (
	"image/color"
	"time"
)

// NightVision is a lasting effect that causes the affected entity to see in dark places as though they were
// fully lit up.
type NightVision struct {
	lastingEffect
}

// WithSettings ...
func (n NightVision) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return NightVision{n.withSettings(d, level, ambient)}
}

// RGBA ...
func (NightVision) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0xa1, A: 0xff}
}
