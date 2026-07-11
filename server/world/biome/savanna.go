package biome

import "image/color"

// Savanna ...
type Savanna struct{}

// Temperature ...
func (Savanna) Temperature() float64 {
	return 1.2
}

// Rainfall ...
func (Savanna) Rainfall() float64 {
	return 0
}

// Depth ...
func (Savanna) Depth() float64 {
	return 0.125
}

// Scale ...
func (Savanna) Scale() float64 {
	return 0.05
}

// WaterColour ...
func (Savanna) WaterColour() color.RGBA {
	return color.RGBA{R: 0x2c, G: 0x8b, B: 0x9c, A: 0xa5}
}

// Tags ...
func (Savanna) Tags() []string {
	return []string{"animal", "monster", "overworld", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (Savanna) String() string {
	return "savanna"
}

// EncodeBiome ...
func (Savanna) EncodeBiome() int {
	return 35
}
