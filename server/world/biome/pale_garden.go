package biome

import "image/color"

// PaleGarden ...
type PaleGarden struct{}

// Temperature ...
func (PaleGarden) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (PaleGarden) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (PaleGarden) Depth() float64 {
	return 0.1
}

// Scale ...
func (PaleGarden) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (PaleGarden) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (PaleGarden) Tags() []string {
	return []string{"monster", "overworld", "pale_garden"}
}

// String ...
func (PaleGarden) String() string {
	return "pale_garden"
}

// EncodeBiome ...
func (PaleGarden) EncodeBiome() int {
	return 193
}
