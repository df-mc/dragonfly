package biome

import "image/color"

// Swamp ...
type Swamp struct{}

// Temperature ...
func (Swamp) Temperature() float64 {
	return 0.8
}

// Rainfall ...
func (Swamp) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (Swamp) Depth() float64 {
	return -0.2
}

// Scale ...
func (Swamp) Scale() float64 {
	return 0.1
}

// WaterColour ...
func (Swamp) WaterColour() color.RGBA {
	return color.RGBA{R: 0x61, G: 0x7b, B: 0x64, A: 0xa5}
}

// Tags ...
func (Swamp) Tags() []string {
	return []string{"animal", "monster", "overworld", "swamp", "spawns_slimes_on_surface", "slime", "swamp_water_huge_mushroom"}
}

// String ...
func (Swamp) String() string {
	return "swampland"
}

// EncodeBiome ...
func (Swamp) EncodeBiome() int {
	return 6
}
