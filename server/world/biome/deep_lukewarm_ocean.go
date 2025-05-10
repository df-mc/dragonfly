package biome

import "image/color"

// DeepLukewarmOcean ...
type DeepLukewarmOcean struct{}

// Temperature ...
func (DeepLukewarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (DeepLukewarmOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (DeepLukewarmOcean) Depth() float64 {
	return -1.8
}

// Scale ...
func (DeepLukewarmOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (DeepLukewarmOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x96, B: 0xdb, A: 0xa5}
}

// Tags ...
func (DeepLukewarmOcean) Tags() []string {
	return []string{"deep", "lukewarm", "monster", "ocean", "overworld", "spawns_warm_variant_farm_animals"}
}

// String ...
func (DeepLukewarmOcean) String() string {
	return "deep_lukewarm_ocean"
}

// EncodeBiome ...
func (DeepLukewarmOcean) EncodeBiome() int {
	return 43
}
