package biome

import "image/color"

// CherryGrove ...
type CherryGrove struct{}

// Temperature ...
func (CherryGrove) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (CherryGrove) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (CherryGrove) Depth() float64 {
	return 0.1
}

// Scale ...
func (CherryGrove) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (CherryGrove) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (CherryGrove) Tags() []string {
	return []string{"mountains", "monster", "overworld", "cherry_grove", "bee_habitat"}
}

// String ...
func (CherryGrove) String() string {
	return "cherry_grove"
}

// EncodeBiome ...
func (CherryGrove) EncodeBiome() int {
	return 192
}
