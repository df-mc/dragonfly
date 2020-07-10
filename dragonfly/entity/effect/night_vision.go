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

// WithDuration ...
func (n NightVision) WithDuration(d time.Duration) entity.Effect {
	return NightVision{n.withDuration(d)}
}

// RGBA ...
func (NightVision) RGBA() color.RGBA {
	return color.RGBA{R: 0x1f, G: 0x1f, B: 0xa1, A: 0xff}
}
