package biome

import "image/color"

// GravellyMountainsPlus ...
type GravellyMountainsPlus struct{}

// Temperature ...
func (GravellyMountainsPlus) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (GravellyMountainsPlus) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (GravellyMountainsPlus) Depth() float64 {
	return 1
}

// Scale ...
func (GravellyMountainsPlus) Scale() float64 {
	return 0.5
}

// WaterColour ...
func (GravellyMountainsPlus) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0e, G: 0x63, B: 0xab, A: 0xa5}
}

// Tags ...
func (GravellyMountainsPlus) Tags() []string {
	return []string{"animal", "extreme_hills", "forest", "monster", "mutated", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (GravellyMountainsPlus) String() string {
	return "extreme_hills_plus_trees_mutated"
}

// EncodeBiome ...
func (GravellyMountainsPlus) EncodeBiome() int {
	return 162
}
