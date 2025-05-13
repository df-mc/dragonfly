package biome

import "image/color"

// ColdOcean ...
type ColdOcean struct{}

// Temperature ...
func (ColdOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (ColdOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (ColdOcean) Depth() float64 {
	return -1
}

// Scale ...
func (ColdOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (ColdOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x80, B: 0xc9, A: 0xa5}
}

// Tags ...
func (ColdOcean) Tags() []string {
	return []string{"cold", "monster", "ocean", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (ColdOcean) String() string {
	return "cold_ocean"
}

// EncodeBiome ...
func (ColdOcean) EncodeBiome() int {
	return 44
}
