package biome

import "image/color"

type DeepOcean struct{}

func (DeepOcean) Temperature() float64 {
	return 0.5
}

func (DeepOcean) Rainfall() float64 {
	return 0.5
}

func (DeepOcean) Depth() float64 {
	return -1.8
}

func (DeepOcean) Scale() float64 {
	return 0.1
}

func (DeepOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x17, G: 0x87, B: 0xd4, A: 0xa5}
}

func (DeepOcean) Tags() []string {
	return []string{"deep", "monster", "ocean", "overworld", "spawns_warm_variant_farm_animals"}
}

func (DeepOcean) String() string {
	return "deep_ocean"
}

func (DeepOcean) EncodeBiome() int {
	return 24
}
