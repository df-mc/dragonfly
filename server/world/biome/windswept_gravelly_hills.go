package biome

import "image/color"

type WindsweptGravellyHills struct{}

func (WindsweptGravellyHills) Temperature() float64 {
	return 0.2
}

func (WindsweptGravellyHills) Rainfall() float64 {
	return 0.3
}

func (WindsweptGravellyHills) Depth() float64 {
	return 1
}

func (WindsweptGravellyHills) Scale() float64 {
	return 0.5
}

func (WindsweptGravellyHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0e, G: 0x63, B: 0xab, A: 0xa5}
}

func (WindsweptGravellyHills) Tags() []string {
	return []string{"animal", "extreme_hills", "monster", "mutated", "overworld", "spawns_cold_variant_farm_animals"}
}

func (WindsweptGravellyHills) String() string {
	return "extreme_hills_mutated"
}

func (WindsweptGravellyHills) EncodeBiome() int {
	return 131
}
