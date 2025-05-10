package biome

import "image/color"

// GiantSpruceTaigaHills ...
type GiantSpruceTaigaHills struct{}

// Temperature ...
func (GiantSpruceTaigaHills) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (GiantSpruceTaigaHills) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (GiantSpruceTaigaHills) Depth() float64 {
	return 0.55
}

// Scale ...
func (GiantSpruceTaigaHills) Scale() float64 {
	return 0.5
}

// WaterColour ...
func (GiantSpruceTaigaHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x28, G: 0x63, B: 0x78, A: 0xa5}
}

// Tags ...
func (GiantSpruceTaigaHills) Tags() []string {
	return []string{"animal", "forest", "hills", "mega", "monster", "mutated", "taiga", "overworld_generation", "spawns_cold_variant_farm_animals"}
}

// String ...
func (GiantSpruceTaigaHills) String() string {
	return "redwood_taiga_hills_mutated"
}

// EncodeBiome ...
func (GiantSpruceTaigaHills) EncodeBiome() int {
	return 161
}
