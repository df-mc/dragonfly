package biome

import "image/color"

// SnowyMountains ...
type SnowyMountains struct{}

// Temperature ...
func (SnowyMountains) Temperature() float64 {
	return 0
}

// Rainfall ...
func (SnowyMountains) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (SnowyMountains) Depth() float64 {
	return 0.45
}

// Scale ...
func (SnowyMountains) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (SnowyMountains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x11, G: 0x56, B: 0xa7, A: 0xa5}
}

// Tags ...
func (SnowyMountains) Tags() []string {
	return []string{"frozen", "ice", "mountain", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (SnowyMountains) String() string {
	return "ice_mountains"
}

// EncodeBiome ...
func (SnowyMountains) EncodeBiome() int {
	return 13
}
