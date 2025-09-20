package biome

import "image/color"

type End struct{}

func (End) Temperature() float64 {
	return 0.5
}

func (End) Rainfall() float64 {
	return 0.5
}

func (End) Depth() float64 {
	return 0.1
}

func (End) Scale() float64 {
	return 0.2
}

func (End) WaterColour() color.RGBA {
	return color.RGBA{R: 0x62, G: 0x52, B: 0x9e, A: 0xa5}
}

func (End) Tags() []string {
	return []string{"the_end", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs"}
}

func (End) String() string {
	return "the_end"
}

func (End) EncodeBiome() int {
	return 9
}
