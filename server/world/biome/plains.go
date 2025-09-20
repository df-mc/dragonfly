package biome

import "image/color"

type Plains struct{}

func (Plains) Temperature() float64 {
	return 0.8
}

func (Plains) Rainfall() float64 {
	return 0.4
}

func (Plains) Depth() float64 {
	return 0.125
}

func (Plains) Scale() float64 {
	return 0.05
}

func (Plains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x44, G: 0xaf, B: 0xf5, A: 0xa5}
}

func (Plains) Tags() []string {
	return []string{"animal", "monster", "overworld", "plains", "bee_habitat"}
}

func (Plains) String() string {
	return "plains"
}

func (Plains) EncodeBiome() int {
	return 1
}
