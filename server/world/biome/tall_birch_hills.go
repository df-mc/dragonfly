package biome

import "image/color"

type TallBirchHills struct{}

func (TallBirchHills) Temperature() float64 {
	return 0.7
}

func (TallBirchHills) Rainfall() float64 {
	return 0.8
}

func (TallBirchHills) Depth() float64 {
	return 0.55
}

func (TallBirchHills) Scale() float64 {
	return 0.5
}

func (TallBirchHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0a, G: 0x74, B: 0xc4, A: 0xa5}
}

func (TallBirchHills) Tags() []string {
	return []string{"animal", "birch", "forest", "hills", "monster", "mutated", "overworld_generation"}
}

func (TallBirchHills) String() string {
	return "birch_forest_hills_mutated"
}

func (TallBirchHills) EncodeBiome() int {
	return 156
}
