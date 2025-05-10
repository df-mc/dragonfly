package biome

import "image/color"

// WarmOcean ...
type WarmOcean struct{}

// Temperature ...
func (WarmOcean) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (WarmOcean) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (WarmOcean) Depth() float64 {
	return -1
}

// Scale ...
func (WarmOcean) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (WarmOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x02, G: 0xb0, B: 0xe5, A: 0xa5}
}

// Tags ...
func (WarmOcean) Tags() []string {
	return []string{"monster", "ocean", "overworld", "warm", "spawns_warm_variant_farm_animals", "spawns_warm_variant_frogs"}
}

// String ...
func (WarmOcean) String() string {
	return "warm_ocean"
}

// EncodeBiome ...
func (WarmOcean) EncodeBiome() int {
	return 40
}
