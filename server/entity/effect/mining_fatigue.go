package effect

import (
	"image/color"
	"math"
)

// MiningFatigue is a lasting effect that decreases the mining speed of a
// player by 10% for each level of the effect.
var MiningFatigue miningFatigue

type miningFatigue struct {
	nopLasting
}

// Multiplier returns the mining speed multiplier from this effect.
func (miningFatigue) Multiplier(lvl int) float64 {
	return math.Pow(3, float64(lvl))
}

// RGBA ...
func (miningFatigue) RGBA() color.RGBA {
	return color.RGBA{R: 0x4a, G: 0x42, B: 0x17, A: 0xff}
}
