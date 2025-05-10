package biome

import "image/color"

// DarkForestHills ...
type DarkForestHills struct{}

// Temperature ...
func (DarkForestHills) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (DarkForestHills) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (DarkForestHills) Depth() float64 {
	return 0.2
}

// Scale ...
func (DarkForestHills) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (DarkForestHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x3b, G: 0x6c, B: 0xd1, A: 0xa5}
}

// Tags ...
func (DarkForestHills) Tags() []string {
	return []string{"animal", "forest", "monster", "mutated", "roofed", "overworld_generation"}
}

// String ...
func (DarkForestHills) String() string {
	return "roofed_forest_mutated"
}

// EncodeBiome ...
func (DarkForestHills) EncodeBiome() int {
	return 157
}
