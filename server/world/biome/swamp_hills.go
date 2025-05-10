package biome

import "image/color"

// SwampHills ...
type SwampHills struct{}

// Temperature ...
func (SwampHills) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (SwampHills) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (SwampHills) Depth() float64 {
	return -0.1
}

// Scale ...
func (SwampHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (SwampHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x4c, G: 0x65, B: 0x59, A: 0xa5}
}

// Tags ...
func (SwampHills) Tags() []string {
	return []string{"animal", "monster", "mutated", "swamp", "overworld_generation", "spawns_slimes_on_surface"}
}

// String ...
func (SwampHills) String() string {
	return "swampland_mutated"
}

// EncodeBiome ...
func (SwampHills) EncodeBiome() int {
	return 134
}
