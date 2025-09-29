package biome

import "image/color"

type OldGrowthBirchForest struct{}

func (OldGrowthBirchForest) Temperature() float64 {
	return 0.7
}

func (OldGrowthBirchForest) Rainfall() float64 {
	return 0.8
}

func (OldGrowthBirchForest) Depth() float64 {
	return 0.2
}

func (OldGrowthBirchForest) Scale() float64 {
	return 0.4
}

func (OldGrowthBirchForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x06, G: 0x77, B: 0xce, A: 0xa5}
}

func (OldGrowthBirchForest) Tags() []string {
	return []string{"animal", "birch", "forest", "monster", "mutated", "bee_habitat", "overworld_generation", "has_structure_trail_ruins"}
}

func (OldGrowthBirchForest) String() string {
	return "birch_forest_mutated"
}

func (OldGrowthBirchForest) EncodeBiome() int {
	return 155
}
