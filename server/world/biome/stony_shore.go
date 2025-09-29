package biome

import "image/color"

type StonyShore struct{}

func (StonyShore) Temperature() float64 {
	return 0.2
}

func (StonyShore) Rainfall() float64 {
	return 0.3
}

func (StonyShore) Depth() float64 {
	return 0.1
}

func (StonyShore) Scale() float64 {
	return 0.8
}

func (StonyShore) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x67, B: 0xbb, A: 0xa5}
}

func (StonyShore) Tags() []string {
	return []string{"beach", "monster", "overworld", "stone"}
}

func (StonyShore) String() string {
	return "stone_beach"
}

func (StonyShore) EncodeBiome() int {
	return 25
}
