package biome

import "image/color"

// DeepDark ...
type DeepDark struct{}

// Temperature ...
func (DeepDark) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (DeepDark) Rainfall() float64 {
	return 0.4
}

// Depth ...
func (DeepDark) Depth() float64 {
	return 0.1
}

// Scale ...
func (DeepDark) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (DeepDark) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (DeepDark) Tags() []string {
	return []string{"caves", "deep_dark", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs"}
}

// String ...
func (DeepDark) String() string {
	return "deep_dark"
}

// EncodeBiome ...
func (DeepDark) EncodeBiome() int {
	return 190
}
