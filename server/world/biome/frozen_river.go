package biome

import "image/color"

type FrozenRiver struct{}

func (FrozenRiver) Temperature() float64 {
	return 0
}

func (FrozenRiver) Rainfall() float64 {
	return 0.5
}

func (FrozenRiver) Depth() float64 {
	return -0.5
}

func (FrozenRiver) Scale() float64 {
	return 0
}

func (FrozenRiver) WaterColour() color.RGBA {
	return color.RGBA{R: 0x18, G: 0x53, B: 0x90, A: 0xa5}
}

func (FrozenRiver) Tags() []string {
	return []string{"frozen", "overworld", "river", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_river_mobs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (FrozenRiver) String() string {
	return "frozen_river"
}

func (FrozenRiver) EncodeBiome() int {
	return 11
}
