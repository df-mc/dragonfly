package biome

import "image/color"

// DeepFrozenOcean ...
type DeepFrozenOcean struct{}

// Temperature ...
func (DeepFrozenOcean) Temperature() float64 {
	return 0
}

// Rainfall ...
func (DeepFrozenOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (DeepFrozenOcean) Depth() float64 {
	return -1.8
}

// Scale ...
func (DeepFrozenOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (DeepFrozenOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x70, B: 0xb5, A: 0xa5}
}

// Tags ...
func (DeepFrozenOcean) Tags() []string {
	return []string{"deep", "frozen", "monster", "ocean", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_polar_bears_on_alternate_blocks"}
}

// String ...
func (DeepFrozenOcean) String() string {
	return "deep_frozen_ocean"
}

// EncodeBiome ...
func (DeepFrozenOcean) EncodeBiome() int {
	return 47
}
