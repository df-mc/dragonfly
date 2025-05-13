package biome

import "image/color"

// LukewarmOcean ...
type LukewarmOcean struct{}

// Temperature ...
func (LukewarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (LukewarmOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (LukewarmOcean) Depth() float64 {
	return -1
}

// Scale ...
func (LukewarmOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (LukewarmOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x96, B: 0xdb, A: 0xa5}
}

// Tags ...
func (LukewarmOcean) Tags() []string {
	return []string{"lukewarm", "monster", "ocean", "overworld", "spawns_warm_variant_farm_animals"}
}

// String ...
func (LukewarmOcean) String() string {
	return "lukewarm_ocean"
}

// EncodeBiome ...
func (LukewarmOcean) EncodeBiome() int {
	return 42
}
