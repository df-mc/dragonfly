package biome

import "image/color"

type DarkForest struct{}

func (DarkForest) Temperature() float64 {
	return 0.7
}

func (DarkForest) Rainfall() float64 {
	return 0.8
}

func (DarkForest) Depth() float64 {
	return 0.1
}

func (DarkForest) Scale() float64 {
	return 0.2
}

func (DarkForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x3b, G: 0x6c, B: 0xd1, A: 0xa5}
}

func (DarkForest) Tags() []string {
	return []string{"animal", "forest", "monster", "no_legacy_worldgen", "overworld", "roofed"}
}

func (DarkForest) String() string {
	return "roofed_forest"
}

func (DarkForest) EncodeBiome() int {
	return 29
}
