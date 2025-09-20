package biome

import "image/color"

type LegacyFrozenOcean struct{}

func (LegacyFrozenOcean) Temperature() float64 {
	return 0
}

func (LegacyFrozenOcean) Rainfall() float64 {
	return 0.5
}

func (LegacyFrozenOcean) Depth() float64 {
	return -1
}

func (LegacyFrozenOcean) Scale() float64 {
	return 0.1
}

func (LegacyFrozenOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (LegacyFrozenOcean) Tags() []string {
	return []string{"legacy", "frozen", "ocean", "overworld", "spawns_cold_variant_farm_animals", "spawns_polar_bears_on_alternate_blocks"}
}

func (LegacyFrozenOcean) String() string {
	return "legacy_frozen_ocean"
}

func (LegacyFrozenOcean) EncodeBiome() int {
	return 10
}
