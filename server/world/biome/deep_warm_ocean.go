package biome

import "image/color"

// DeepWarmOcean ...
type DeepWarmOcean struct{}

// Temperature ...
func (DeepWarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepWarmOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (DeepWarmOcean) Depth() float64 {
	return -1.8
}

// Scale ...
func (DeepWarmOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (DeepWarmOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x02, G: 0xb0, B: 0xe5, A: 0xa5}
}

// Tags ...
func (DeepWarmOcean) Tags() []string {
	return []string{"deep", "monster", "ocean", "overworld", "warm", "spawns_warm_variant_farm_animals"}
}

// String ...
func (DeepWarmOcean) String() string {
	return "deep_warm_ocean"
}

// EncodeBiome ...
func (DeepWarmOcean) EncodeBiome() int {
	return 41
}
