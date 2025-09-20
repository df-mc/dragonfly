package biome

import "image/color"

type WindsweptForest struct{}

func (WindsweptForest) Temperature() float64 {
	return 0.2
}

func (WindsweptForest) Rainfall() float64 {
	return 0.3
}

func (WindsweptForest) Depth() float64 {
	return 1
}

func (WindsweptForest) Scale() float64 {
	return 0.5
}

func (WindsweptForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0e, G: 0x63, B: 0xab, A: 0xa5}
}

func (WindsweptForest) Tags() []string {
	return []string{"animal", "extreme_hills", "forest", "monster", "mountain", "overworld", "spawns_cold_variant_farm_animals"}
}

func (WindsweptForest) String() string {
	return "extreme_hills_plus_trees"
}

func (WindsweptForest) EncodeBiome() int {
	return 34
}
