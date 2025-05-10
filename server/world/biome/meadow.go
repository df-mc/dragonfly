package biome

import "image/color"

// Meadow ...
type Meadow struct{}

// Temperature ...
func (Meadow) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (Meadow) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (Meadow) Depth() float64 {
	return 0.1
}

// Scale ...
func (Meadow) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (Meadow) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (Meadow) Tags() []string {
	return []string{"mountains", "monster", "overworld", "meadow", "bee_habitat"}
}

// String ...
func (Meadow) String() string {
	return "meadow"
}

// EncodeBiome ...
func (Meadow) EncodeBiome() int {
	return 186
}
