package biome

import "image/color"

type BirchForest struct{}

func (BirchForest) Temperature() float64 {
	return 0.6
}

func (BirchForest) Rainfall() float64 {
	return 0.6
}

func (BirchForest) Depth() float64 {
	return 0.1
}

func (BirchForest) Scale() float64 {
	return 0.2
}

func (BirchForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x06, G: 0x77, B: 0xce, A: 0xa5}
}

func (BirchForest) Tags() []string {
	return []string{"animal", "birch", "forest", "monster", "overworld", "bee_habitat"}
}

func (BirchForest) String() string {
	return "birch_forest"
}

func (BirchForest) EncodeBiome() int {
	return 27
}
