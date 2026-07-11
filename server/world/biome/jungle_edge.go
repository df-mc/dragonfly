package biome

import "image/color"

// JungleEdge ...
type JungleEdge struct{}

// Temperature ...
func (JungleEdge) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (JungleEdge) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (JungleEdge) Depth() float64 {
	return 0.1
}

// Scale ...
func (JungleEdge) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (JungleEdge) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x8a, B: 0xe3, A: 0xa5}
}

// Tags ...
func (JungleEdge) Tags() []string {
	return []string{"animal", "edge", "jungle", "monster", "overworld", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (JungleEdge) String() string {
	return "jungle_edge"
}

// EncodeBiome ...
func (JungleEdge) EncodeBiome() int {
	return 23
}
