package biome

import "image/color"

// WindsweptForest ...
type WindsweptForest struct{}

// Temperature ...
func (WindsweptForest) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (WindsweptForest) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (WindsweptForest) Depth() float64 {
	return 1
}

// Scale ...
func (WindsweptForest) Scale() float64 {
	return 0.5
}

// WaterColour ...
func (WindsweptForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0e, G: 0x63, B: 0xab, A: 0xa5}
}

// Tags ...
func (WindsweptForest) Tags() []string {
	return []string{"animal", "extreme_hills", "forest", "monster", "mountain", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (WindsweptForest) String() string {
	return "extreme_hills_plus_trees"
}

// EncodeBiome ...
func (WindsweptForest) EncodeBiome() int {
	return 34
}
