package biome

import "image/color"

// WindsweptSavanna ...
type WindsweptSavanna struct{}

// Temperature ...
func (WindsweptSavanna) Temperature() float64 {
	return 1.1
}

// Rainfall ...
func (WindsweptSavanna) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (WindsweptSavanna) Depth() float64 {
	return 0.363
}

// Scale ...
func (WindsweptSavanna) Scale() float64 {
	return 1.225
}

// WaterColour ...
func (WindsweptSavanna) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (WindsweptSavanna) Tags() []string {
	return []string{"animal", "monster", "mutated", "overworld", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (WindsweptSavanna) String() string {
	return "savanna_mutated"
}

// EncodeBiome ...
func (WindsweptSavanna) EncodeBiome() int {
	return 163
}
