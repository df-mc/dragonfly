package biome

import "image/color"

// Plains ...
type Plains struct{}

// Temperature ...
func (Plains) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Plains) Rainfall() float64 {
	return 0.4
}

// Depth ...
func (Plains) Depth() float64 {
	return 0.125
}

// Scale ...
func (Plains) Scale() float64 {
	return 0.05
}

// WaterColour ...
func (Plains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x44, G: 0xaf, B: 0xf5, A: 0xa5}
}

// Tags ...
func (Plains) Tags() []string {
	return []string{"animal", "monster", "overworld", "plains", "bee_habitat"}
}

// String ...
func (Plains) String() string {
	return "plains"
}

// EncodeBiome ...
func (Plains) EncodeBiome() int {
	return 1
}
