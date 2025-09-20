package biome

import "image/color"

type SunflowerPlains struct{}

func (SunflowerPlains) Temperature() float64 {
	return 0.8
}

func (SunflowerPlains) Rainfall() float64 {
	return 0.4
}

func (SunflowerPlains) Depth() float64 {
	return 0.125
}

func (SunflowerPlains) Scale() float64 {
	return 0.05
}

func (SunflowerPlains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (SunflowerPlains) Tags() []string {
	return []string{"animal", "monster", "mutated", "overworld", "plains", "bee_habitat"}
}

func (SunflowerPlains) String() string {
	return "sunflower_plains"
}

func (SunflowerPlains) EncodeBiome() int {
	return 129
}
