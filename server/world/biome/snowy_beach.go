package biome

import "image/color"

type SnowyBeach struct{}

func (SnowyBeach) Temperature() float64 {
	return 0.05
}

func (SnowyBeach) Rainfall() float64 {
	return 0.3
}

func (SnowyBeach) Depth() float64 {
	return 0
}

func (SnowyBeach) Scale() float64 {
	return 0.025
}

func (SnowyBeach) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0x63, B: 0xa5, A: 0xa5}
}

func (SnowyBeach) Tags() []string {
	return []string{"beach", "cold", "monster", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (SnowyBeach) String() string {
	return "cold_beach"
}

func (SnowyBeach) EncodeBiome() int {
	return 26
}
