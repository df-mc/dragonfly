package biome

import "image/color"

// StonyPeaks ...
type StonyPeaks struct{}

// Temperature ...
func (StonyPeaks) Temperature() float64 {
	return 1
}

// Rainfall ...
func (StonyPeaks) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (StonyPeaks) Depth() float64 {
	return 0.1
}

// Scale ...
func (StonyPeaks) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (StonyPeaks) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (StonyPeaks) Tags() []string {
	return []string{"mountains", "monster", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (StonyPeaks) String() string {
	return "stony_peaks"
}

// EncodeBiome ...
func (StonyPeaks) EncodeBiome() int {
	return 189
}
