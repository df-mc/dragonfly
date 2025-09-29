package biome

import "image/color"

type Ocean struct{}

func (Ocean) Temperature() float64 {
	return 0.5
}

func (Ocean) Rainfall() float64 {
	return 0.5
}

func (Ocean) Depth() float64 {
	return -1
}

func (Ocean) Scale() float64 {
	return 0.1
}

func (Ocean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x17, G: 0x87, B: 0xd4, A: 0xa5}
}

func (Ocean) Tags() []string {
	return []string{"monster", "ocean", "overworld"}
}

func (Ocean) String() string {
	return "ocean"
}

func (Ocean) EncodeBiome() int {
	return 0
}
