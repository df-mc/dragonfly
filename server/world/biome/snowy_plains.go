package biome

import "image/color"

// SnowyPlains ...
type SnowyPlains struct{}

// Temperature ...
func (SnowyPlains) Temperature() float64 {
	return 0
}

// Rainfall ...
func (SnowyPlains) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (SnowyPlains) Depth() float64 {
	return 0.125
}

// Scale ...
func (SnowyPlains) Scale() float64 {
	return 0.05
}

// WaterColour ...
func (SnowyPlains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0x55, B: 0x9b, A: 0xa5}
}

// Tags ...
func (SnowyPlains) Tags() []string {
	return []string{"frozen", "ice", "ice_plains", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (SnowyPlains) String() string {
	return "ice_plains"
}

// EncodeBiome ...
func (SnowyPlains) EncodeBiome() int {
	return 12
}
