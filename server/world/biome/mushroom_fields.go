package biome

import "image/color"

// MushroomFields ...
type MushroomFields struct{}

// Temperature ...
func (MushroomFields) Temperature() float64 {
	return 0.9
}

// Rainfall ...
func (MushroomFields) Rainfall() float64 {
	return 1
}

// Depth ...
func (MushroomFields) Depth() float64 {
	return 0.2
}

// Scale ...
func (MushroomFields) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (MushroomFields) WaterColour() color.RGBA {
	return color.RGBA{R: 0x8a, G: 0x89, B: 0x97, A: 0xa5}
}

// Tags ...
func (MushroomFields) Tags() []string {
	return []string{"mooshroom_island", "overworld", "spawns_without_patrols"}
}

// String ...
func (MushroomFields) String() string {
	return "mushroom_island"
}

// EncodeBiome ...
func (MushroomFields) EncodeBiome() int {
	return 14
}
