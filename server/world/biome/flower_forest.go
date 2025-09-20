package biome

import "image/color"

type FlowerForest struct{}

func (FlowerForest) Temperature() float64 {
	return 0.7
}

func (FlowerForest) Rainfall() float64 {
	return 0.8
}

func (FlowerForest) Depth() float64 {
	return 0.1
}

func (FlowerForest) Scale() float64 {
	return 0.4
}

func (FlowerForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0xa3, B: 0xcc, A: 0xa5}
}

func (FlowerForest) Tags() []string {
	return []string{"animal", "flower_forest", "monster", "mutated", "overworld", "bee_habitat"}
}

func (FlowerForest) String() string {
	return "flower_forest"
}

func (FlowerForest) EncodeBiome() int {
	return 132
}
