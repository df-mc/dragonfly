package biome

import "image/color"

type DripstoneCaves struct{}

func (DripstoneCaves) Temperature() float64 {
	return 0.2
}

func (DripstoneCaves) Rainfall() float64 {
	return 0
}

func (DripstoneCaves) Depth() float64 {
	return 0.1
}

func (DripstoneCaves) Scale() float64 {
	return 0.2
}

func (DripstoneCaves) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (DripstoneCaves) Tags() []string {
	return []string{"caves", "overworld", "dripstone_caves", "monster"}
}

func (DripstoneCaves) String() string {
	return "dripstone_caves"
}

func (DripstoneCaves) EncodeBiome() int {
	return 188
}
