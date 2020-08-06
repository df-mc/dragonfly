package effect

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/damage"
	"image/color"
	"time"
)

// Resistance is a lasting effect that reduces the damage taken from any sources except for void damage or
// custom damage.
type Resistance struct {
	lastingEffect
}

// Multiplier returns a damage multiplier for the damage source passed.
func (r Resistance) Multiplier(e damage.Source) float64 {
	switch e.(type) {
	case damage.SourceVoid, damage.SourceStarvation, damage.SourceCustom:
		return 1
	}
	v := 1 - 0.2*float64(r.Lvl)
	if v <= 0 {
		v = 0
	}
	return v
}

// WithSettings ...
func (r Resistance) WithSettings(d time.Duration, level int, ambient bool) Effect {
	return Resistance{r.withSettings(d, level, ambient)}
}

// RGBA ...
func (Resistance) RGBA() color.RGBA {
	return color.RGBA{R: 0x99, G: 0x45, B: 0x3a, A: 0xff}
}
