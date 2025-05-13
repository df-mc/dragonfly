package biome

import "image/color"

// FrozenOcean ...
type FrozenOcean struct{}

// Temperature ...
func (FrozenOcean) Temperature() float64 {
	return 0
}

// Rainfall ...
func (FrozenOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (FrozenOcean) Depth() float64 {
	return -1
}

// Scale ...
func (FrozenOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (FrozenOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x70, B: 0xb5, A: 0xa5}
}

// Tags ...
func (FrozenOcean) Tags() []string {
	return []string{"frozen", "monster", "ocean", "overworld", "spawns_polar_bears_on_alternate_blocks", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (FrozenOcean) String() string {
	return "frozen_ocean"
}

// EncodeBiome ...
func (FrozenOcean) EncodeBiome() int {
	return 46
}
