package biome

import "image/color"

// MountainEdge ...
type MountainEdge struct{}

// Temperature ...
func (MountainEdge) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (MountainEdge) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (MountainEdge) Depth() float64 {
	return 0.8
}

// Scale ...
func (MountainEdge) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (MountainEdge) WaterColour() color.RGBA {
	return color.RGBA{R: 0x04, G: 0x5c, B: 0xd5, A: 0xa5}
}

// Tags ...
func (MountainEdge) Tags() []string {
	return []string{"animal", "edge", "extreme_hills", "monster", "mountain", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (MountainEdge) String() string {
	return "extreme_hills_edge"
}

// EncodeBiome ...
func (MountainEdge) EncodeBiome() int {
	return 20
}
