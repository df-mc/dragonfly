package biome

import "image/color"

// TallBirchHills ...
type TallBirchHills struct{}

// Temperature ...
func (TallBirchHills) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (TallBirchHills) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (TallBirchHills) Depth() float64 {
	return 0.55
}

// Scale ...
func (TallBirchHills) Scale() float64 {
	return 0.5
}

// WaterColour ...
func (TallBirchHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0a, G: 0x74, B: 0xc4, A: 0xa5}
}

// Tags ...
func (TallBirchHills) Tags() []string {
	return []string{"animal", "birch", "forest", "hills", "monster", "mutated", "overworld_generation"}
}

// String ...
func (TallBirchHills) String() string {
	return "birch_forest_hills_mutated"
}

// EncodeBiome ...
func (TallBirchHills) EncodeBiome() int {
	return 156
}
