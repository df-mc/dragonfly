package biome

import "image/color"

// TaigaMountains ...
type TaigaMountains struct{}

// Temperature ...
func (TaigaMountains) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (TaigaMountains) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (TaigaMountains) Depth() float64 {
	return 0.2
}

// Scale ...
func (TaigaMountains) Scale() float64 {
	return 0.4
}

// WaterColour ...
func (TaigaMountains) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1e, G: 0x6b, B: 0x82, A: 0xa5}
}

// Tags ...
func (TaigaMountains) Tags() []string {
	return []string{"animal", "forest", "monster", "mutated", "taiga", "overworld_generation", "spawns_cold_variant_farm_animals"}
}

// String ...
func (TaigaMountains) String() string {
	return "taiga_mutated"
}

// EncodeBiome ...
func (TaigaMountains) EncodeBiome() int {
	return 133
}
