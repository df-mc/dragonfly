package biome

import "image/color"

type ErodedBadlands struct{}

func (ErodedBadlands) Temperature() float64 {
	return 2
}

func (ErodedBadlands) Rainfall() float64 {
	return 0
}

func (ErodedBadlands) Depth() float64 {
	return 0.1
}

func (ErodedBadlands) Scale() float64 {
	return 0.2
}

func (ErodedBadlands) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0xa2, B: 0xc5, A: 0xa5}
}

func (ErodedBadlands) Tags() []string {
	return []string{"animal", "mesa", "monster", "mutated", "overworld", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

func (ErodedBadlands) String() string {
	return "mesa_bryce"
}

func (ErodedBadlands) EncodeBiome() int {
	return 165
}
