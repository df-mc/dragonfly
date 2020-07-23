package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// NightVision is a lasting effect that causes the affected entity to see in dark places as though they were
// fully lit up.
type NightVision struct {
	lastingEffect
}

// WithDurationAndLevel ...
func (n NightVision) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return NightVision{n.withDurationAndLevel(d, level)}
}

// RGBA ...
func (NightVision) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0xa1, A: 0xff}
}
