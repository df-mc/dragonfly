package biome

import "image/color"

type Jungle struct{}

func (Jungle) Temperature() float64 {
	return 0.95
}

func (Jungle) Rainfall() float64 {
	return 0.9
}

func (Jungle) Depth() float64 {
	return 0.1
}

func (Jungle) Scale() float64 {
	return 0.2
}

func (Jungle) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0xa2, B: 0xc5, A: 0xa5}
}

func (Jungle) Tags() []string {
	return []string{"animal", "has_structure_trail_ruins", "jungle", "monster", "overworld", "rare", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

func (Jungle) String() string {
	return "jungle"
}

func (Jungle) EncodeBiome() int {
	return 21
}
