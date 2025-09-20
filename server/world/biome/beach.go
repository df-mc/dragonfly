package biome

import "image/color"

type Beach struct{}

func (Beach) Temperature() float64 {
	return 0.8
}

func (Beach) Rainfall() float64 {
	return 0.4
}

func (Beach) Depth() float64 {
	return 0
}

func (Beach) Scale() float64 {
	return 0.025
}

func (Beach) WaterColour() color.RGBA {
	return color.RGBA{R: 0x15, G: 0x7c, B: 0xab, A: 0xa5}
}

func (Beach) Tags() []string {
	return []string{"beach", "monster", "overworld", "warm"}
}

func (Beach) String() string {
	return "beach"
}

func (Beach) EncodeBiome() int {
	return 16
}
