package biome

import "image/color"

type ModifiedJungle struct{}

func (ModifiedJungle) Temperature() float64 {
	return 0.95
}

func (ModifiedJungle) Rainfall() float64 {
	return 0.9
}

func (ModifiedJungle) Depth() float64 {
	return 0.2
}

func (ModifiedJungle) Scale() float64 {
	return 0.4
}

func (ModifiedJungle) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1b, G: 0x9e, B: 0xd8, A: 0xa5}
}

func (ModifiedJungle) Tags() []string {
	return []string{"animal", "jungle", "monster", "mutated", "overworld_generation", "spawns_warm_variant_farm_animals"}
}

func (ModifiedJungle) String() string {
	return "jungle_mutated"
}

func (ModifiedJungle) EncodeBiome() int {
	return 149
}
