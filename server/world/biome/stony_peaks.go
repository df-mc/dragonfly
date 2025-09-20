package biome

import "image/color"

type StonyPeaks struct{}

func (StonyPeaks) Temperature() float64 {
	return 1
}

func (StonyPeaks) Rainfall() float64 {
	return 0.3
}

func (StonyPeaks) Depth() float64 {
	return 0.1
}

func (StonyPeaks) Scale() float64 {
	return 0.2
}

func (StonyPeaks) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (StonyPeaks) Tags() []string {
	return []string{"mountains", "monster", "overworld", "spawns_cold_variant_farm_animals"}
}

func (StonyPeaks) String() string {
	return "stony_peaks"
}

func (StonyPeaks) EncodeBiome() int {
	return 189
}
