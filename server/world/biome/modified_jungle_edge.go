package biome

import "image/color"

type ModifiedJungleEdge struct{}

func (ModifiedJungleEdge) Temperature() float64 {
	return 0.95
}

func (ModifiedJungleEdge) Rainfall() float64 {
	return 0.8
}

func (ModifiedJungleEdge) Depth() float64 {
	return 0.2
}

func (ModifiedJungleEdge) Scale() float64 {
	return 0.4
}

func (ModifiedJungleEdge) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x8a, B: 0xe3, A: 0xa5}
}

func (ModifiedJungleEdge) Tags() []string {
	return []string{"animal", "edge", "jungle", "monster", "mutated", "overworld_generation", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

func (ModifiedJungleEdge) String() string {
	return "jungle_edge_mutated"
}

func (ModifiedJungleEdge) EncodeBiome() int {
	return 151
}
