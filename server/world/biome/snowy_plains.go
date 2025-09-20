package biome

import "image/color"

type SnowyPlains struct{}

func (SnowyPlains) Temperature() float64 {
	return 0
}

func (SnowyPlains) Rainfall() float64 {
	return 0.5
}

func (SnowyPlains) Depth() float64 {
	return 0.125
}

func (SnowyPlains) Scale() float64 {
	return 0.05
}

func (SnowyPlains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0x55, B: 0x9b, A: 0xa5}
}

func (SnowyPlains) Tags() []string {
	return []string{"frozen", "ice", "ice_plains", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (SnowyPlains) String() string {
	return "ice_plains"
}

func (SnowyPlains) EncodeBiome() int {
	return 12
}
