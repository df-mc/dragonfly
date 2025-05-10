package biome

import "image/color"

// JaggedPeaks ...
type JaggedPeaks struct{}

// Temperature ...
func (JaggedPeaks) Temperature() float64 {
	return -0.7
}

// Rainfall ...
func (JaggedPeaks) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (JaggedPeaks) Depth() float64 {
	return 0.1
}

// Scale ...
func (JaggedPeaks) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (JaggedPeaks) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (JaggedPeaks) Tags() []string {
	return []string{"mountains", "monster", "overworld", "frozen", "jagged_peaks", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

// String ...
func (JaggedPeaks) String() string {
	return "jagged_peaks"
}

// EncodeBiome ...
func (JaggedPeaks) EncodeBiome() int {
	return 182
}
