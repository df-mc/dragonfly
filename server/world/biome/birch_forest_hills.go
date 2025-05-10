package biome

import "image/color"

// BirchForestHills ...
type BirchForestHills struct{}

// Temperature ...
func (BirchForestHills) Temperature() float64 {
	return 0.6
}

// Rainfall ...
func (BirchForestHills) Rainfall() float64 {
	return 0.6
}

// Depth ...
func (BirchForestHills) Depth() float64 {
	return 0.45
}

// Scale ...
func (BirchForestHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (BirchForestHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0a, G: 0x74, B: 0xc4, A: 0xa5}
}

// Tags ...
func (BirchForestHills) Tags() []string {
	return []string{"animal", "birch", "forest", "hills", "monster", "overworld", "bee_habitat"}
}

// String ...
func (BirchForestHills) String() string {
	return "birch_forest_hills"
}

// EncodeBiome ...
func (BirchForestHills) EncodeBiome() int {
	return 28
}
