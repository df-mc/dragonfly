package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity"
	"image/color"
	"time"
)

// JumpBoost is a lasting effect that causes the affected entity to be able to jump much higher, depending on
// the level of the effect.
type JumpBoost struct {
	lastingEffect
}

// WithDuration ...
func (j JumpBoost) WithDuration(d time.Duration) entity.Effect {
	return JumpBoost{j.withDuration(d)}
}

// RGBA ...
func (JumpBoost) RGBA() color.RGBA {
	return color.RGBA{R: 0x22, G: 0xff, B: 0x4c, A: 0xff}
}
