package biome

import "image/color"

type DeepDark struct{}

func (DeepDark) Temperature() float64 {
	return 0.8
}

func (DeepDark) Rainfall() float64 {
	return 0.4
}

func (DeepDark) Depth() float64 {
	return 0.1
}

func (DeepDark) Scale() float64 {
	return 0.2
}

func (DeepDark) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (DeepDark) Tags() []string {
	return []string{"caves", "deep_dark", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs"}
}

func (DeepDark) String() string {
	return "deep_dark"
}

func (DeepDark) EncodeBiome() int {
	return 190
}
