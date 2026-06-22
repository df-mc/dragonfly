package biome

import "image/color"

// WarpedForest ...
type WarpedForest struct{}

// Temperature ...
func (WarpedForest) Temperature() float64 {
	return 2
}

// Rainfall ...
func (WarpedForest) Rainfall() float64 {
	return 0
}

// Depth ...
func (WarpedForest) Depth() float64 {
	return 0.1
}

// Scale ...
func (WarpedForest) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (WarpedForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

// Tags ...
func (WarpedForest) Tags() []string {
	return []string{"nether", "netherwart_forest", "warped_forest", "spawn_endermen", "spawns_warm_variant_farm_animals"}
}

// String ...
func (WarpedForest) String() string {
	return "warped_forest"
}

// EncodeBiome ...
func (WarpedForest) EncodeBiome() int {
	return 180
}
