package biome

import "image/color"

type WindsweptHills struct{}

func (WindsweptHills) Temperature() float64 {
	return 0.2
}

func (WindsweptHills) Rainfall() float64 {
	return 0.3
}

func (WindsweptHills) Depth() float64 {
	return 1
}

func (WindsweptHills) Scale() float64 {
	return 0.5
}

func (WindsweptHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x00, G: 0x7b, B: 0xf7, A: 0xa5}
}

func (WindsweptHills) Tags() []string {
	return []string{"animal", "extreme_hills", "monster", "overworld", "spawns_cold_variant_farm_animals"}
}

func (WindsweptHills) String() string {
	return "extreme_hills"
}

func (WindsweptHills) EncodeBiome() int {
	return 3
}
