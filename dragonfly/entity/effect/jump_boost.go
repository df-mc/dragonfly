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

// WithDurationAndLevel ...
func (j JumpBoost) WithDurationAndLevel(d time.Duration, level int) entity.Effect {
	return JumpBoost{j.withDurationAndLevel(d, level)}
}

// RGBA ...
func (JumpBoost) RGBA() color.RGBA {
	return color.RGBA{R: 0x22, G: 0xff, B: 0x4c, A: 0xff}
}
