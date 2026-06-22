package biome

import "image/color"

// CrimsonForest ...
type CrimsonForest struct{}

// Temperature ...
func (CrimsonForest) Temperature() float64 {
	return 2
}

// Rainfall ...
func (CrimsonForest) Rainfall() float64 {
	return 0
}

// Depth ...
func (CrimsonForest) Depth() float64 {
	return 0.1
}

// Scale ...
func (CrimsonForest) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (CrimsonForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

// Tags ...
func (CrimsonForest) Tags() []string {
	return []string{"nether", "netherwart_forest", "crimson_forest", "spawn_few_zombified_piglins", "spawn_piglin", "spawns_warm_variant_farm_animals"}
}

// String ...
func (CrimsonForest) String() string {
	return "crimson_forest"
}

// EncodeBiome ...
func (CrimsonForest) EncodeBiome() int {
	return 179
}
