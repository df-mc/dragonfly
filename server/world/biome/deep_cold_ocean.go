package biome

import "image/color"

type DeepColdOcean struct{}

func (DeepColdOcean) Temperature() float64 {
	return 0.5
}

func (DeepColdOcean) Rainfall() float64 {
	return 0.5
}

func (DeepColdOcean) Depth() float64 {
	return -1.8
}

func (DeepColdOcean) Scale() float64 {
	return 0.1
}

func (DeepColdOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x80, B: 0xc9, A: 0xa5}
}

func (DeepColdOcean) Tags() []string {
	return []string{"cold", "deep", "monster", "ocean", "overworld", "spawns_cold_variant_farm_animals"}
}

func (DeepColdOcean) String() string {
	return "deep_cold_ocean"
}

func (DeepColdOcean) EncodeBiome() int {
	return 45
}
