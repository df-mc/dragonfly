package biome

import "image/color"

type GiantSpruceTaigaHills struct{}

func (GiantSpruceTaigaHills) Temperature() float64 {
	return 0.3
}

func (GiantSpruceTaigaHills) Rainfall() float64 {
	return 0.8
}

func (GiantSpruceTaigaHills) Depth() float64 {
	return 0.55
}

func (GiantSpruceTaigaHills) Scale() float64 {
	return 0.5
}

func (GiantSpruceTaigaHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x28, G: 0x63, B: 0x78, A: 0xa5}
}

func (GiantSpruceTaigaHills) Tags() []string {
	return []string{"animal", "forest", "hills", "mega", "monster", "mutated", "taiga", "overworld_generation", "spawns_cold_variant_farm_animals"}
}

func (GiantSpruceTaigaHills) String() string {
	return "redwood_taiga_hills_mutated"
}

func (GiantSpruceTaigaHills) EncodeBiome() int {
	return 161
}
