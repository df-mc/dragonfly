package biome

import "image/color"

type Swamp struct{}

func (Swamp) Temperature() float64 {
	return 0.8
}

func (Swamp) Rainfall() float64 {
	return 0.5
}

func (Swamp) Depth() float64 {
	return -0.2
}

func (Swamp) Scale() float64 {
	return 0.1
}

func (Swamp) WaterColour() color.RGBA {
	return color.RGBA{R: 0x4c, G: 0x65, B: 0x59, A: 0xa5}
}

func (Swamp) Tags() []string {
	return []string{"animal", "monster", "overworld", "swamp", "spawns_slimes_on_surface"}
}

func (Swamp) String() string {
	return "swampland"
}

func (Swamp) EncodeBiome() int {
	return 6
}
