package biome

import "image/color"

type WarpedForest struct{}

func (WarpedForest) Temperature() float64 {
	return 2
}

func (WarpedForest) Rainfall() float64 {
	return 0
}

func (WarpedForest) Spores() (blueSpores float64, redSpores float64) {
	return 0.25, 0
}

func (WarpedForest) Depth() float64 {
	return 0.1
}

func (WarpedForest) Scale() float64 {
	return 0.2
}

func (WarpedForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

func (WarpedForest) Tags() []string {
	return []string{"nether", "netherwart_forest", "warped_forest", "spawn_endermen", "spawns_warm_variant_farm_animals"}
}

func (WarpedForest) String() string {
	return "warped_forest"
}

func (WarpedForest) EncodeBiome() int {
	return 180
}
