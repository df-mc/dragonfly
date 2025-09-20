package biome

import "image/color"

type CrimsonForest struct{}

func (CrimsonForest) Temperature() float64 {
	return 2
}

func (CrimsonForest) Rainfall() float64 {
	return 0
}

func (CrimsonForest) Spores() (blueSpores float64, redSpores float64) {
	return 0, 0.25
}

func (CrimsonForest) Depth() float64 {
	return 0.1
}

func (CrimsonForest) Scale() float64 {
	return 0.2
}

func (CrimsonForest) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

func (CrimsonForest) Tags() []string {
	return []string{"nether", "netherwart_forest", "crimson_forest", "spawn_few_zombified_piglins", "spawn_piglin", "spawns_warm_variant_farm_animals"}
}

func (CrimsonForest) String() string {
	return "crimson_forest"
}

func (CrimsonForest) EncodeBiome() int {
	return 179
}
