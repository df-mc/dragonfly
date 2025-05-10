package biome

import "image/color"

// LegacyFrozenOcean ...
type LegacyFrozenOcean struct{}

// Temperature ...
func (LegacyFrozenOcean) Temperature() float64 {
	return 0
}

// Rainfall ...
func (LegacyFrozenOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (LegacyFrozenOcean) Depth() float64 {
	return -1
}

// Scale ...
func (LegacyFrozenOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (LegacyFrozenOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (LegacyFrozenOcean) Tags() []string {
	return []string{"legacy", "frozen", "ocean", "overworld", "spawns_cold_variant_farm_animals", "spawns_polar_bears_on_alternate_blocks"}
}

// String ...
func (LegacyFrozenOcean) String() string {
	return "legacy_frozen_ocean"
}

// EncodeBiome ...
func (LegacyFrozenOcean) EncodeBiome() int {
	return 10
}
