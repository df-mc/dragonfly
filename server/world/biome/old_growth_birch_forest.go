package biome

import "image/color"

// OldGrowthBirchForest ...
type OldGrowthBirchForest struct{}

// Temperature ...
func (OldGrowthBirchForest) Temperature() float64 {
	return 0.7
}

// Rainfall ...
func (OldGrowthBirchForest) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (OldGrowthBirchForest) Depth() float64 {
	return 0.2
}

// Scale ...
func (OldGrowthBirchForest) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (OldGrowthBirchForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x06, G: 0x77, B: 0xce, A: 0xa5}
}

// Tags ...
func (OldGrowthBirchForest) Tags() []string {
	return []string{"animal", "birch", "forest", "monster", "mutated", "bee_habitat", "overworld_generation", "has_structure_trail_ruins"}
}

// String ...
func (OldGrowthBirchForest) String() string {
	return "birch_forest_mutated"
}

// EncodeBiome ...
func (OldGrowthBirchForest) EncodeBiome() int {
	return 155
}
