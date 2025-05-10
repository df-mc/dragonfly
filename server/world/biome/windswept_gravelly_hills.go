package biome

import "image/color"

// WindsweptGravellyHills ...
type WindsweptGravellyHills struct{}

// Temperature ...
func (WindsweptGravellyHills) Temperature() float64 {
	return 0.2
}

// Rainfall ...
func (WindsweptGravellyHills) Rainfall() float64 {
	return 0.3
}

// Depth ...
func (WindsweptGravellyHills) Depth() float64 {
	return 1
}

// Scale ...
func (WindsweptGravellyHills) Scale() float64 {
	return 0.5
}

// WaterColour ...
func (WindsweptGravellyHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0e, G: 0x63, B: 0xab, A: 0xa5}
}

// Tags ...
func (WindsweptGravellyHills) Tags() []string {
	return []string{"animal", "extreme_hills", "monster", "mutated", "overworld", "spawns_cold_variant_farm_animals"}
}

// String ...
func (WindsweptGravellyHills) String() string {
	return "extreme_hills_mutated"
}

// EncodeBiome ...
func (WindsweptGravellyHills) EncodeBiome() int {
	return 131
}
