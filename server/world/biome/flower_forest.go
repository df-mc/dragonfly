package biome

import "image/color"

// FlowerForest ...
type FlowerForest struct{}

// Temperature ...
func (FlowerForest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (FlowerForest) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (FlowerForest) Depth() float64 {
	return 0.1
}

// Scale ...
func (FlowerForest) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (FlowerForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0xa3, B: 0xcc, A: 0xa5}
}

// Tags ...
func (FlowerForest) Tags() []string {
	return []string{"animal", "flower_forest", "monster", "mutated", "overworld", "bee_habitat"}
}

// String ...
func (FlowerForest) String() string {
	return "flower_forest"
}

// EncodeBiome ...
func (FlowerForest) EncodeBiome() int {
	return 132
}
