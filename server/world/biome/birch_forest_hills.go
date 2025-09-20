package biome

import "image/color"

type BirchForestHills struct{}

func (BirchForestHills) Temperature() float64 {
	return 0.6
}

func (BirchForestHills) Rainfall() float64 {
	return 0.6
}

func (BirchForestHills) Depth() float64 {
	return 0.45
}

func (BirchForestHills) Scale() float64 {
	return 0.3
}

func (BirchForestHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0a, G: 0x74, B: 0xc4, A: 0xa5}
}

func (BirchForestHills) Tags() []string {
	return []string{"animal", "birch", "forest", "hills", "monster", "overworld", "bee_habitat"}
}

func (BirchForestHills) String() string {
	return "birch_forest_hills"
}

func (BirchForestHills) EncodeBiome() int {
	return 28
}
