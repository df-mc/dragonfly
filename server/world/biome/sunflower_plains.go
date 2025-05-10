package biome

import "image/color"

// SunflowerPlains ...
type SunflowerPlains struct{}

// Temperature ...
func (SunflowerPlains) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (SunflowerPlains) Rainfall() float64 {
	return 0.4
}

// Depth ...
func (SunflowerPlains) Depth() float64 {
	return 0.125
}

// Scale ...
func (SunflowerPlains) Scale() float64 {
	return 0.05
}

// WaterColour ...
func (SunflowerPlains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (SunflowerPlains) Tags() []string {
	return []string{"animal", "monster", "mutated", "overworld", "plains", "bee_habitat"}
}

// String ...
func (SunflowerPlains) String() string {
	return "sunflower_plains"
}

// EncodeBiome ...
func (SunflowerPlains) EncodeBiome() int {
	return 129
}
