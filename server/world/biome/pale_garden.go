package biome

import "image/color"

type PaleGarden struct{}

func (PaleGarden) Temperature() float64 {
	return 0.7
}

func (PaleGarden) Rainfall() float64 {
	return 0.8
}

func (PaleGarden) Depth() float64 {
	return 0.1
}

func (PaleGarden) Scale() float64 {
	return 0.2
}

func (PaleGarden) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (PaleGarden) Tags() []string {
	return []string{"monster", "overworld", "pale_garden"}
}

func (PaleGarden) String() string {
	return "pale_garden"
}

func (PaleGarden) EncodeBiome() int {
	return 193
}
