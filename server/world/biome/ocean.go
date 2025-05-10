package biome

import "image/color"

// Ocean ...
type Ocean struct{}

// Temperature ...
func (Ocean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (Ocean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (Ocean) Depth() float64 {
	return -1
}

// Scale ...
func (Ocean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (Ocean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x17, G: 0x87, B: 0xd4, A: 0xa5}
}

// Tags ...
func (Ocean) Tags() []string {
	return []string{"monster", "ocean", "overworld"}
}

// String ...
func (Ocean) String() string {
	return "ocean"
}

// EncodeBiome ...
func (Ocean) EncodeBiome() int {
	return 0
}
