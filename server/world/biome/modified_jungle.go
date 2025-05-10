package biome

import "image/color"

// ModifiedJungle ...
type ModifiedJungle struct{}

// Temperature ...
func (ModifiedJungle) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (ModifiedJungle) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (ModifiedJungle) Depth() float64 {
	return 0.2
}

// Scale ...
func (ModifiedJungle) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (ModifiedJungle) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1b, G: 0x9e, B: 0xd8, A: 0xa5}
}

// Tags ...
func (ModifiedJungle) Tags() []string {
	return []string{"animal", "jungle", "monster", "mutated", "overworld_generation", "spawns_warm_variant_farm_animals"}
}

// String ...
func (ModifiedJungle) String() string {
	return "jungle_mutated"
}

// EncodeBiome ...
func (ModifiedJungle) EncodeBiome() int {
	return 149
}
