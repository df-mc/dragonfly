package biome

import "image/color"

type TaigaHills struct{}

func (TaigaHills) Temperature() float64 {
	return 0.25
}

func (TaigaHills) Rainfall() float64 {
	return 0.8
}

func (TaigaHills) Depth() float64 {
	return 0.45
}

func (TaigaHills) Scale() float64 {
	return 0.3
}

func (TaigaHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x23, G: 0x65, B: 0x83, A: 0xa5}
}

func (TaigaHills) Tags() []string {
	return []string{"animal", "hills", "monster", "overworld", "forest", "taiga", "spawns_cold_variant_farm_animals"}
}

func (TaigaHills) String() string {
	return "taiga_hills"
}

func (TaigaHills) EncodeBiome() int {
	return 19
}
