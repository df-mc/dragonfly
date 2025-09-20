package biome

import "image/color"

type SwampHills struct{}

func (SwampHills) Temperature() float64 {
	return 0.8
}

func (SwampHills) Rainfall() float64 {
	return 0.5
}

func (SwampHills) Depth() float64 {
	return -0.1
}

func (SwampHills) Scale() float64 {
	return 0.3
}

func (SwampHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x4c, G: 0x65, B: 0x59, A: 0xa5}
}

func (SwampHills) Tags() []string {
	return []string{"animal", "monster", "mutated", "swamp", "overworld_generation", "spawns_slimes_on_surface"}
}

func (SwampHills) String() string {
	return "swampland_mutated"
}

func (SwampHills) EncodeBiome() int {
	return 134
}
