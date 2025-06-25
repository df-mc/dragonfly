package biome

import "image/color"

// BirchForest ...
type BirchForest struct{}

// Temperature ...
func (BirchForest) Temperature() float64 {
	return 0.6
}

// Rainfall ...
func (BirchForest) Rainfall() float64 {
	return 0.6
}

// Depth ...
func (BirchForest) Depth() float64 {
	return 0.1
}

// Scale ...
func (BirchForest) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (BirchForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x06, G: 0x77, B: 0xce, A: 0xa5}
}

// Tags ...
func (BirchForest) Tags() []string {
	return []string{"animal", "birch", "forest", "monster", "overworld", "bee_habitat"}
}

// String ...
func (BirchForest) String() string {
	return "birch_forest"
}

// EncodeBiome ...
func (BirchForest) EncodeBiome() int {
	return 27
}
