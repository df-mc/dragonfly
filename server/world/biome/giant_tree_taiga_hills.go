package biome

import "image/color"

type GiantTreeTaigaHills struct{}

func (GiantTreeTaigaHills) Temperature() float64 {
	return 0.3
}

func (GiantTreeTaigaHills) Rainfall() float64 {
	return 0.8
}

func (GiantTreeTaigaHills) Depth() float64 {
	return 0.45
}

func (GiantTreeTaigaHills) Scale() float64 {
	return 0.3
}

func (GiantTreeTaigaHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x28, G: 0x63, B: 0x78, A: 0xa5}
}

func (GiantTreeTaigaHills) Tags() []string {
	return []string{"animal", "forest", "hills", "mega", "monster", "overworld", "taiga", "spawns_cold_variant_farm_animals"}
}

func (GiantTreeTaigaHills) String() string {
	return "mega_taiga_hills"
}

func (GiantTreeTaigaHills) EncodeBiome() int {
	return 33
}
