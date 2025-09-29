package biome

import "image/color"

type Forest struct{}

func (Forest) Temperature() float64 {
	return 0.7
}

func (Forest) Rainfall() float64 {
	return 0.8
}

func (Forest) Depth() float64 {
	return 0.1
}

func (Forest) Scale() float64 {
	return 0.2
}

func (Forest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1e, G: 0x97, B: 0xf2, A: 0xa5}
}

func (Forest) Tags() []string {
	return []string{"animal", "forest", "monster", "overworld", "bee_habitat"}
}

func (Forest) String() string {
	return "forest"
}

func (Forest) EncodeBiome() int {
	return 4
}
