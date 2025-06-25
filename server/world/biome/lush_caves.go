package biome

import "image/color"

// LushCaves ...
type LushCaves struct{}

// Temperature ...
func (LushCaves) Temperature() float64 {
	return 0.9
}

// Rainfall ...
func (LushCaves) Rainfall() float64 {
	return 0
}

// Depth ...
func (LushCaves) Depth() float64 {
	return 0.1
}

// Scale ...
func (LushCaves) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (LushCaves) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (LushCaves) Tags() []string {
	return []string{"caves", "lush_caves", "overworld", "monster", "spawns_tropical_fish_at_any_height"}
}

// String ...
func (LushCaves) String() string {
	return "lush_caves"
}

// EncodeBiome ...
func (LushCaves) EncodeBiome() int {
	return 187
}
