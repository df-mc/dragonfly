package biome

import "image/color"

// SulfurCaves ...
type SulfurCaves struct{}

// Temperature ...
func (SulfurCaves) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (SulfurCaves) Rainfall() float64 {
	return 0.4
}

// Depth ...
func (SulfurCaves) Depth() float64 {
	return 0.1
}

// Scale ...
func (SulfurCaves) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (SulfurCaves) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (SulfurCaves) Tags() []string {
	return []string{"caves", "sulfur_caves", "overworld", "monster"}
}

// String ...
func (SulfurCaves) String() string {
	return "sulfur_caves"
}

// EncodeBiome ...
func (SulfurCaves) EncodeBiome() int {
	return 194
}
