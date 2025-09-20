package biome

import "image/color"

type Badlands struct{}

func (Badlands) Temperature() float64 {
	return 2
}

func (Badlands) Rainfall() float64 {
	return 0
}

func (Badlands) Depth() float64 {
	return 0.1
}

func (Badlands) Scale() float64 {
	return 0.2
}

func (Badlands) WaterColour() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x7f, B: 0x81, A: 0xa5}
}

func (Badlands) Tags() []string {
	return []string{"animal", "mesa", "monster", "overworld", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

func (Badlands) String() string {
	return "mesa"
}

func (Badlands) EncodeBiome() int {
	return 37
}
