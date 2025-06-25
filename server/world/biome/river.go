package biome

import "image/color"

// River ...
type River struct{}

// Temperature ...
func (River) Temperature() float64 {
	return 0.5
}

// Rainfall ...
func (River) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (River) Depth() float64 {
	return -0.5
}

// Scale ...
func (River) Scale() float64 {
	return 0
}

// WaterColour ...
func (River) WaterColour() color.RGBA {
	return color.RGBA{R: 0x00, G: 0x84, B: 0xff, A: 0xa5}
}

// Tags ...
func (River) Tags() []string {
	return []string{"overworld", "spawns_more_frequent_drowned", "spawns_reduced_water_ambient_mobs", "spawns_river_mobs", "river"}
}

// String ...
func (River) String() string {
	return "river"
}

// EncodeBiome ...
func (River) EncodeBiome() int {
	return 7
}
