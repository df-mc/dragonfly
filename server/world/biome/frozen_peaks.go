package biome

import "image/color"

type FrozenPeaks struct{}

func (FrozenPeaks) Temperature() float64 {
	return -0.7
}

func (FrozenPeaks) Rainfall() float64 {
	return 0.9
}

func (FrozenPeaks) Depth() float64 {
	return 0.1
}

func (FrozenPeaks) Scale() float64 {
	return 0.2
}

func (FrozenPeaks) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (FrozenPeaks) Tags() []string {
	return []string{"mountains", "monster", "overworld", "frozen", "frozen_peaks", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (FrozenPeaks) String() string {
	return "frozen_peaks"
}

func (FrozenPeaks) EncodeBiome() int {
	return 183
}
