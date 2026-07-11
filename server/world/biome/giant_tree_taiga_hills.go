package biome

import "image/color"

// GiantTreeTaigaHills ...
type GiantTreeTaigaHills struct{}

// Temperature ...
func (GiantTreeTaigaHills) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (GiantTreeTaigaHills) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (GiantTreeTaigaHills) Depth() float64 {
	return 0.45
}

// Scale ...
func (GiantTreeTaigaHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (GiantTreeTaigaHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x28, G: 0x63, B: 0x78, A: 0xa5}
}

// Tags ...
func (GiantTreeTaigaHills) Tags() []string {
	return []string{"animal", "forest", "hills", "mega", "monster", "overworld", "taiga", "spawns_cold_variant_farm_animals"}
}

// String ...
func (GiantTreeTaigaHills) String() string {
	return "mega_taiga_hills"
}

// EncodeBiome ...
func (GiantTreeTaigaHills) EncodeBiome() int {
	return 33
}
