package biome

import "image/color"

type LushCaves struct{}

func (LushCaves) Temperature() float64 {
	return 0.9
}

func (LushCaves) Rainfall() float64 {
	return 0
}

func (LushCaves) Depth() float64 {
	return 0.1
}

func (LushCaves) Scale() float64 {
	return 0.2
}

func (LushCaves) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (LushCaves) Tags() []string {
	return []string{"caves", "lush_caves", "overworld", "monster", "spawns_tropical_fish_at_any_height"}
}

func (LushCaves) String() string {
	return "lush_caves"
}

func (LushCaves) EncodeBiome() int {
	return 187
}
