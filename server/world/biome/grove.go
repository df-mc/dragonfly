package biome

import "image/color"

// Grove ...
type Grove struct{}

// Temperature ...
func (Grove) Temperature() float64 {
	return -0.2
}

// Rainfall ...
func (Grove) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (Grove) Depth() float64 {
	return 0.1
}

// Scale ...
func (Grove) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (Grove) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (Grove) Tags() []string {
	return []string{"mountains", "cold", "monster", "overworld", "grove", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (Grove) String() string {
	return "grove"
}

// EncodeBiome ...
func (Grove) EncodeBiome() int {
	return 185
}
