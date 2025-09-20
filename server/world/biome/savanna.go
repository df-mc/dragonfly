package biome

import "image/color"

type Savanna struct{}

func (Savanna) Temperature() float64 {
	return 1.2
}

func (Savanna) Rainfall() float64 {
	return 0
}

func (Savanna) Depth() float64 {
	return 0.125
}

func (Savanna) Scale() float64 {
	return 0.05
}

func (Savanna) WaterColour() color.RGBA {
	return color.RGBA{R: 0x2c, G: 0x8b, B: 0x9c, A: 0xa5}
}

func (Savanna) Tags() []string {
	return []string{"animal", "monster", "overworld", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

func (Savanna) String() string {
	return "savanna"
}

func (Savanna) EncodeBiome() int {
	return 35
}
