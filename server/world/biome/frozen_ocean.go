package biome

import "image/color"

type FrozenOcean struct{}

func (FrozenOcean) Temperature() float64 {
	return 0
}

func (FrozenOcean) Rainfall() float64 {
	return 0.5
}

func (FrozenOcean) Depth() float64 {
	return -1
}

func (FrozenOcean) Scale() float64 {
	return 0.1
}

func (FrozenOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x70, B: 0xb5, A: 0xa5}
}

func (FrozenOcean) Tags() []string {
	return []string{"frozen", "monster", "ocean", "overworld", "spawns_polar_bears_on_alternate_blocks", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (FrozenOcean) String() string {
	return "frozen_ocean"
}

func (FrozenOcean) EncodeBiome() int {
	return 46
}
