package biome

import "image/color"

// DarkForest ...
type DarkForest struct{}

// Temperature ...
func (DarkForest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (DarkForest) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (DarkForest) Depth() float64 {
	return 0.1
}

// Scale ...
func (DarkForest) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (DarkForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x3b, G: 0x6c, B: 0xd1, A: 0xa5}
}

// Tags ...
func (DarkForest) Tags() []string {
	return []string{"animal", "forest", "monster", "no_legacy_worldgen", "overworld", "roofed"}
}

// String ...
func (DarkForest) String() string {
	return "roofed_forest"
}

// EncodeBiome ...
func (DarkForest) EncodeBiome() int {
	return 29
}
