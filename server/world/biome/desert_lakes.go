package biome

import "image/color"

type DesertLakes struct{}

func (DesertLakes) Temperature() float64 {
	return 2
}

func (DesertLakes) Rainfall() float64 {
	return 0
}

func (DesertLakes) Depth() float64 {
	return 0.225
}

func (DesertLakes) Scale() float64 {
	return 0.25
}

func (DesertLakes) WaterColour() color.RGBA {
	return color.RGBA{R: 0x32, G: 0xa5, B: 0x98, A: 0xa5}
}

func (DesertLakes) Tags() []string {
	return []string{"desert", "monster", "mutated", "overworld_generation", "spawns_gold_rabbits", "spawns_warm_variant_farm_animals"}
}

func (DesertLakes) String() string {
	return "desert_mutated"
}

func (DesertLakes) EncodeBiome() int {
	return 130
}
