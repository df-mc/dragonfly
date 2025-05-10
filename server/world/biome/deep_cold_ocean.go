package biome

import "image/color"

// DeepColdOcean ...
type DeepColdOcean struct{}

// Temperature ...
func (DeepColdOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepColdOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (DeepColdOcean) Depth() float64 {
	return -1.8
}

// Scale ...
func (DeepColdOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (DeepColdOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x80, B: 0xc9, A: 0xa5}
}

// Tags ...
func (DeepColdOcean) Tags() []string {
	return []string{"cold", "deep", "monster", "ocean", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (DeepColdOcean) String() string {
	return "deep_cold_ocean"
}

// EncodeBiome ...
func (DeepColdOcean) EncodeBiome() int {
	return 45
}
