package biome

import "image/color"

// OldGrowthSpruceTaiga ...
type OldGrowthSpruceTaiga struct{}

// Temperature ...
func (OldGrowthSpruceTaiga) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (OldGrowthSpruceTaiga) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (OldGrowthSpruceTaiga) Depth() float64 {
	return 0.2
}

// Scale ...
func (OldGrowthSpruceTaiga) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (OldGrowthSpruceTaiga) WaterColour() color.RGBA {
	return color.RGBA{R: 0x2d, G: 0x6d, B: 0x77, A: 0xa5}
}

// Tags ...
func (OldGrowthSpruceTaiga) Tags() []string {
	return []string{"animal", "forest", "mega", "monster", "mutated", "overworld", "taiga", "has_structure_trail_ruins", "spawns_cold_variant_farm_animals"}
}

// String ...
func (OldGrowthSpruceTaiga) String() string {
	return "redwood_taiga_mutated"
}

// EncodeBiome ...
func (OldGrowthSpruceTaiga) EncodeBiome() int {
	return 160
}
