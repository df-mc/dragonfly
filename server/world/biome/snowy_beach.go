package biome

import "image/color"

// SnowyBeach ...
type SnowyBeach struct{}

// Temperature ...
func (SnowyBeach) Temperature() float64 {
	return 0.05
}

// Rainfall ...
func (SnowyBeach) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (SnowyBeach) Depth() float64 {
	return 0
}

// Scale ...
func (SnowyBeach) Scale() float64 {
	return 0.025
}

// WaterColour ...
func (SnowyBeach) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0x63, B: 0xa5, A: 0xa5}
}

// Tags ...
func (SnowyBeach) Tags() []string {
	return []string{"beach", "cold", "monster", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (SnowyBeach) String() string {
	return "cold_beach"
}

// EncodeBiome ...
func (SnowyBeach) EncodeBiome() int {
	return 26
}
