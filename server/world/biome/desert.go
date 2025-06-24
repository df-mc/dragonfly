package biome

import "image/color"

// Desert ...
type Desert struct{}

// Temperature ...
func (Desert) Temperature() float64 {
	return 2
}

// Rainfall ...
func (Desert) Rainfall() float64 {
	return 0
}

// Depth ...
func (Desert) Depth() float64 {
	return 0.125
}

// Scale ...
func (Desert) Scale() float64 {
	return 0.05
}

// WaterColour ...
func (Desert) WaterColour() color.RGBA {
	return color.RGBA{R: 0x32, G: 0xa5, B: 0x98, A: 0xa5}
}

// Tags ...
func (Desert) Tags() []string {
	return []string{"desert", "monster", "overworld", "spawns_gold_rabbits", "spawns_warm_variant_farm_animals", "spawns_warm_variant_frogs"}
}

// String ...
func (Desert) String() string {
	return "desert"
}

// EncodeBiome ...
func (Desert) EncodeBiome() int {
	return 2
}
