package biome

import "image/color"

// FrozenPeaks ...
type FrozenPeaks struct{}

// Temperature ...
func (FrozenPeaks) Temperature() float64 {
	return -0.7
}

// Rainfall ...
func (FrozenPeaks) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (FrozenPeaks) Depth() float64 {
	return 0.1
}

// Scale ...
func (FrozenPeaks) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (FrozenPeaks) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (FrozenPeaks) Tags() []string {
	return []string{"mountains", "monster", "overworld", "frozen", "frozen_peaks", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (FrozenPeaks) String() string {
	return "frozen_peaks"
}

// EncodeBiome ...
func (FrozenPeaks) EncodeBiome() int {
	return 183
}
