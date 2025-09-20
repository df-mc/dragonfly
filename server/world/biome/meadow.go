package biome

import "image/color"

type Meadow struct{}

func (Meadow) Temperature() float64 {
	return 0.3
}

func (Meadow) Rainfall() float64 {
	return 0.8
}

func (Meadow) Depth() float64 {
	return 0.1
}

func (Meadow) Scale() float64 {
	return 0.2
}

func (Meadow) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (Meadow) Tags() []string {
	return []string{"mountains", "monster", "overworld", "meadow", "bee_habitat"}
}

func (Meadow) String() string {
	return "meadow"
}

func (Meadow) EncodeBiome() int {
	return 186
}
