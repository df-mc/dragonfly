package biome

import "image/color"

// ErodedBadlands ...
type ErodedBadlands struct{}

// Temperature ...
func (ErodedBadlands) Temperature() float64 {
	return 2
}

// Rainfall ...
func (ErodedBadlands) Rainfall() float64 {
	return 0
}

// Depth ...
func (ErodedBadlands) Depth() float64 {
	return 0.1
}

// Scale ...
func (ErodedBadlands) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (ErodedBadlands) WaterColour() color.RGBA {
	return color.RGBA{R: 0x14, G: 0xa2, B: 0xc5, A: 0xa5}
}

// Tags ...
func (ErodedBadlands) Tags() []string {
	return []string{"animal", "mesa", "monster", "mutated", "overworld", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (ErodedBadlands) String() string {
	return "mesa_bryce"
}

// EncodeBiome ...
func (ErodedBadlands) EncodeBiome() int {
	return 165
}
