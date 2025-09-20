package biome

import "image/color"

type SnowyMountains struct{}

func (SnowyMountains) Temperature() float64 {
	return 0
}

func (SnowyMountains) Rainfall() float64 {
	return 0.5
}

func (SnowyMountains) Depth() float64 {
	return 0.45
}

func (SnowyMountains) Scale() float64 {
	return 0.3
}

func (SnowyMountains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x11, G: 0x56, B: 0xa7, A: 0xa5}
}

func (SnowyMountains) Tags() []string {
	return []string{"frozen", "ice", "mountain", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (SnowyMountains) String() string {
	return "ice_mountains"
}

func (SnowyMountains) EncodeBiome() int {
	return 13
}
