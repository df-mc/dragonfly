package biome

import "image/color"

// MushroomFieldShore ...
type MushroomFieldShore struct{}

// Temperature ...
func (MushroomFieldShore) Temperature() float64 {
	return 0.9
}

// Rainfall ...
func (MushroomFieldShore) Rainfall() float64 {
	return 1
}

// Depth ...
func (MushroomFieldShore) Depth() float64 {
	return 0
}

// Scale ...
func (MushroomFieldShore) Scale() float64 {
	return 0.025
}

// WaterColour ...
func (MushroomFieldShore) WaterColour() color.RGBA {
	return color.RGBA{R: 0x81, G: 0x81, B: 0x93, A: 0xa5}
}

// Tags ...
func (MushroomFieldShore) Tags() []string {
	return []string{"mooshroom_island", "overworld", "shore", "spawns_without_patrols"}
}

// String ...
func (MushroomFieldShore) String() string {
	return "mushroom_island_shore"
}

// EncodeBiome ...
func (MushroomFieldShore) EncodeBiome() int {
	return 15
}
