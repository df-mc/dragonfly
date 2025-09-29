package biome

import "image/color"

type JungleEdge struct{}

func (JungleEdge) Temperature() float64 {
	return 0.95
}

func (JungleEdge) Rainfall() float64 {
	return 0.8
}

func (JungleEdge) Depth() float64 {
	return 0.1
}

func (JungleEdge) Scale() float64 {
	return 0.2
}

func (JungleEdge) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x8a, B: 0xe3, A: 0xa5}
}

func (JungleEdge) Tags() []string {
	return []string{"animal", "edge", "jungle", "monster", "overworld", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

func (JungleEdge) String() string {
	return "jungle_edge"
}

func (JungleEdge) EncodeBiome() int {
	return 23
}
