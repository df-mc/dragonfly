package biome

import "image/color"

// TaigaHills ...
type TaigaHills struct{}

// Temperature ...
func (TaigaHills) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (TaigaHills) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (TaigaHills) Depth() float64 {
	return 0.45
}

// Scale ...
func (TaigaHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (TaigaHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x23, G: 0x65, B: 0x83, A: 0xa5}
}

// Tags ...
func (TaigaHills) Tags() []string {
	return []string{"animal", "hills", "monster", "overworld", "forest", "taiga", "spawns_cold_variant_farm_animals"}
}

// String ...
func (TaigaHills) String() string {
	return "taiga_hills"
}

// EncodeBiome ...
func (TaigaHills) EncodeBiome() int {
	return 19
}
