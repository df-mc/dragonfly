package biome

import "image/color"

// Badlands ...
type Badlands struct{}

// Temperature ...
func (Badlands) Temperature() float64 {
	return 2
}

// Rainfall ...
func (Badlands) Rainfall() float64 {
	return 0
}

// Depth ...
func (Badlands) Depth() float64 {
	return 0.1
}

// Scale ...
func (Badlands) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (Badlands) WaterColour() color.RGBA {
	return color.RGBA{R: 0x4e, G: 0x7f, B: 0x81, A: 0xa5}
}

// Tags ...
func (Badlands) Tags() []string {
	return []string{"animal", "mesa", "monster", "overworld", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (Badlands) String() string {
	return "mesa"
}

// EncodeBiome ...
func (Badlands) EncodeBiome() int {
	return 37
}
