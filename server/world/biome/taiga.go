package biome

import "image/color"

// Taiga ...
type Taiga struct{}

// Temperature ...
func (Taiga) Temperature() float64 {
	return 0.25
}

// Rainfall ...
func (Taiga) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (Taiga) Depth() float64 {
	return 0.1
}

// Scale ...
func (Taiga) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (Taiga) WaterColour() color.RGBA {
	return color.RGBA{R: 0x28, G: 0x70, B: 0x82, A: 0xa5}
}

// Tags ...
func (Taiga) Tags() []string {
	return []string{"animal", "forest", "monster", "overworld", "taiga", "has_structure_trail_ruins", "spawns_cold_variant_farm_animals"}
}

// String ...
func (Taiga) String() string {
	return "taiga"
}

// EncodeBiome ...
func (Taiga) EncodeBiome() int {
	return 5
}
