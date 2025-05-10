package biome

import "image/color"

// StonyShore ...
type StonyShore struct{}

// Temperature ...
func (StonyShore) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (StonyShore) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (StonyShore) Depth() float64 {
	return 0.1
}

// Scale ...
func (StonyShore) Scale() float64 {
	return 0.8
}

// WaterColour ...
func (StonyShore) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x67, B: 0xbb, A: 0xa5}
}

// Tags ...
func (StonyShore) Tags() []string {
	return []string{"beach", "monster", "overworld", "stone"}
}

// String ...
func (StonyShore) String() string {
	return "stone_beach"
}

// EncodeBiome ...
func (StonyShore) EncodeBiome() int {
	return 25
}
