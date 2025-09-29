package biome

import "image/color"

type DarkForestHills struct{}

func (DarkForestHills) Temperature() float64 {
	return 0.7
}

func (DarkForestHills) Rainfall() float64 {
	return 0.8
}

func (DarkForestHills) Depth() float64 {
	return 0.2
}

func (DarkForestHills) Scale() float64 {
	return 0.4
}

func (DarkForestHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x3b, G: 0x6c, B: 0xd1, A: 0xa5}
}

func (DarkForestHills) Tags() []string {
	return []string{"animal", "forest", "monster", "mutated", "roofed", "overworld_generation"}
}

func (DarkForestHills) String() string {
	return "roofed_forest_mutated"
}

func (DarkForestHills) EncodeBiome() int {
	return 157
}
