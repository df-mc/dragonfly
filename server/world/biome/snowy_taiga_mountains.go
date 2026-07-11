package biome

import "image/color"

// SnowyTaigaMountains ...
type SnowyTaigaMountains struct{}

// Temperature ...
func (SnowyTaigaMountains) Temperature() float64 {
	return -0.5
}

// Rainfall ...
func (SnowyTaigaMountains) Rainfall() float64 {
	return 0.4
}

// Depth ...
func (SnowyTaigaMountains) Depth() float64 {
	return 0.3
}

// Scale ...
func (SnowyTaigaMountains) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (SnowyTaigaMountains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x5e, B: 0x83, A: 0xa5}
}

// Tags ...
func (SnowyTaigaMountains) Tags() []string {
	return []string{"animal", "cold", "forest", "monster", "mutated", "taiga", "overworld_generation", "spawns_cold_variant_farm_animals"}
}

// String ...
func (SnowyTaigaMountains) String() string {
	return "cold_taiga_mutated"
}

// EncodeBiome ...
func (SnowyTaigaMountains) EncodeBiome() int {
	return 158
}
