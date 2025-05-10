package biome

import "image/color"

// WoodedHills ...
type WoodedHills struct{}

// Temperature ...
func (WoodedHills) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (WoodedHills) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (WoodedHills) Depth() float64 {
	return 0.45
}

// Scale ...
func (WoodedHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (WoodedHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x05, G: 0x6b, B: 0xd1, A: 0xa5}
}

// Tags ...
func (WoodedHills) Tags() []string {
	return []string{"animal", "hills", "monster", "overworld", "forest", "bee_habitat"}
}

// String ...
func (WoodedHills) String() string {
	return "forest_hills"
}

// EncodeBiome ...
func (WoodedHills) EncodeBiome() int {
	return 18
}
