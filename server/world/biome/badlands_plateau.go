package biome

import "image/color"

// BadlandsPlateau ...
type BadlandsPlateau struct{}

// Temperature ...
func (BadlandsPlateau) Temperature() float64 {
	return 2
}

// Rainfall ...
func (BadlandsPlateau) Rainfall() float64 {
	return 0
}

// Depth ...
func (BadlandsPlateau) Depth() float64 {
	return 1.5
}

// Scale ...
func (BadlandsPlateau) Scale() float64 {
	return 0.025
}

// WaterColour ...
func (BadlandsPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x80, B: 0x9e, A: 0xa5}
}

// Tags ...
func (BadlandsPlateau) Tags() []string {
	return []string{"animal", "mesa", "monster", "overworld", "plateau", "rare", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (BadlandsPlateau) String() string {
	return "mesa_plateau"
}

// EncodeBiome ...
func (BadlandsPlateau) EncodeBiome() int {
	return 39
}
