package biome

import "image/color"

type Desert struct{}

func (Desert) Temperature() float64 {
	return 2
}

func (Desert) Rainfall() float64 {
	return 0
}

func (Desert) Depth() float64 {
	return 0.125
}

func (Desert) Scale() float64 {
	return 0.05
}

func (Desert) WaterColour() color.RGBA {
	return color.RGBA{R: 0x32, G: 0xa5, B: 0x98, A: 0xa5}
}

func (Desert) Tags() []string {
	return []string{"desert", "monster", "overworld", "spawns_gold_rabbits", "spawns_warm_variant_farm_animals", "spawns_warm_variant_frogs"}
}

func (Desert) String() string {
	return "desert"
}

func (Desert) EncodeBiome() int {
	return 2
}
