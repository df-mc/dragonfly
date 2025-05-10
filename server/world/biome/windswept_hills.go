package biome

import "image/color"

// WindsweptHills ...
type WindsweptHills struct{}

// Temperature ...
func (WindsweptHills) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (WindsweptHills) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (WindsweptHills) Depth() float64 {
	return 1
}

// Scale ...
func (WindsweptHills) Scale() float64 {
	return 0.5
}

// WaterColour ...
func (WindsweptHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x00, G: 0x7b, B: 0xf7, A: 0xa5}
}

// Tags ...
func (WindsweptHills) Tags() []string {
	return []string{"animal", "extreme_hills", "monster", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (WindsweptHills) String() string {
	return "extreme_hills"
}

// EncodeBiome ...
func (WindsweptHills) EncodeBiome() int {
	return 3
}
