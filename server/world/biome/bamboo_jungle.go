package biome

import "image/color"

// BambooJungle ...
type BambooJungle struct{}

// Temperature ...
func (BambooJungle) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (BambooJungle) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (BambooJungle) Depth() float64 {
	return 0.1
}

// Scale ...
func (BambooJungle) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (BambooJungle) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0xa2, B: 0xc5, A: 0xa5}
}

// Tags ...
func (BambooJungle) Tags() []string {
	return []string{"animal", "bamboo", "jungle", "monster", "overworld", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (BambooJungle) String() string {
	return "bamboo_jungle"
}

// EncodeBiome ...
func (BambooJungle) EncodeBiome() int {
	return 48
}
