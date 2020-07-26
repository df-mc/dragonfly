package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// Nausea is a lasting effect that causes the screen to warp, similarly to when entering a nether portal.
type Nausea struct {
	lastingEffect
}

// WithSettings ...
func (n Nausea) WithSettings(d time.Duration, level int, ambient bool) entity.Effect {
	return Nausea{n.withSettings(d, level, ambient)}
}

// RGBA ...
func (Nausea) RGBA() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x1d, B: 0x4a, A: 0xff}
}
