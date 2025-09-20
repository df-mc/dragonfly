package biome

import "image/color"

type DeepFrozenOcean struct{}

func (DeepFrozenOcean) Temperature() float64 {
	return 0
}

func (DeepFrozenOcean) Rainfall() float64 {
	return 0.5
}

func (DeepFrozenOcean) Depth() float64 {
	return -1.8
}

func (DeepFrozenOcean) Scale() float64 {
	return 0.1
}

func (DeepFrozenOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x70, B: 0xb5, A: 0xa5}
}

func (DeepFrozenOcean) Tags() []string {
	return []string{"deep", "frozen", "monster", "ocean", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_polar_bears_on_alternate_blocks"}
}

func (DeepFrozenOcean) String() string {
	return "deep_frozen_ocean"
}

func (DeepFrozenOcean) EncodeBiome() int {
	return 47
}
