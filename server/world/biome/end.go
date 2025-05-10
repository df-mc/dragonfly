package biome

import "image/color"

// End ...
type End struct{}

// Temperature ...
func (End) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (End) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (End) Depth() float64 {
	return 0.1
}

// Scale ...
func (End) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (End) WaterColour() color.RGBA {
	return color.RGBA{R: 0x62, G: 0x52, B: 0x9e, A: 0xa5}
}

// Tags ...
func (End) Tags() []string {
	return []string{"the_end", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs"}
}

// String ...
func (End) String() string {
	return "the_end"
}

// EncodeBiome ...
func (End) EncodeBiome() int {
	return 9
}
