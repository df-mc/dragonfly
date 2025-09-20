package biome

import "image/color"

type ColdOcean struct{}

func (ColdOcean) Temperature() float64 {
	return 0.5
}

func (ColdOcean) Rainfall() float64 {
	return 0.5
}

func (ColdOcean) Depth() float64 {
	return -1
}

func (ColdOcean) Scale() float64 {
	return 0.1
}

func (ColdOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x80, B: 0xc9, A: 0xa5}
}

func (ColdOcean) Tags() []string {
	return []string{"cold", "monster", "ocean", "overworld", "spawns_cold_variant_farm_animals"}
}

func (ColdOcean) String() string {
	return "cold_ocean"
}

func (ColdOcean) EncodeBiome() int {
	return 44
}
