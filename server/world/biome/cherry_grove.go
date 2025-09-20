package biome

import "image/color"

type CherryGrove struct{}

func (CherryGrove) Temperature() float64 {
	return 0.3
}

func (CherryGrove) Rainfall() float64 {
	return 0.8
}

func (CherryGrove) Depth() float64 {
	return 0.1
}

func (CherryGrove) Scale() float64 {
	return 0.2
}

func (CherryGrove) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (CherryGrove) Tags() []string {
	return []string{"mountains", "monster", "overworld", "cherry_grove", "bee_habitat"}
}

func (CherryGrove) String() string {
	return "cherry_grove"
}

func (CherryGrove) EncodeBiome() int {
	return 192
}
