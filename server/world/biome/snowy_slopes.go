package biome

import "image/color"

type SnowySlopes struct{}

func (SnowySlopes) Temperature() float64 {
	return -0.3
}

func (SnowySlopes) Rainfall() float64 {
	return 0.9
}

func (SnowySlopes) Depth() float64 {
	return 0.1
}

func (SnowySlopes) Scale() float64 {
	return 0.2
}

func (SnowySlopes) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (SnowySlopes) Tags() []string {
	return []string{"mountains", "monster", "overworld", "frozen", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits", "snowy_slopes", "spawns_cold_variant_farm_animals"}
}

func (SnowySlopes) String() string {
	return "snowy_slopes"
}

func (SnowySlopes) EncodeBiome() int {
	return 184
}
