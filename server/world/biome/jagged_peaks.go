package biome

import "image/color"

type JaggedPeaks struct{}

func (JaggedPeaks) Temperature() float64 {
	return -0.7
}

func (JaggedPeaks) Rainfall() float64 {
	return 0.9
}

func (JaggedPeaks) Depth() float64 {
	return 0.1
}

func (JaggedPeaks) Scale() float64 {
	return 0.2
}

func (JaggedPeaks) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (JaggedPeaks) Tags() []string {
	return []string{"mountains", "monster", "overworld", "frozen", "jagged_peaks", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_snow_foxes", "spawns_white_rabbits"}
}

func (JaggedPeaks) String() string {
	return "jagged_peaks"
}

func (JaggedPeaks) EncodeBiome() int {
	return 182
}
