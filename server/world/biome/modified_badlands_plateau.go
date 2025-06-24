package biome

import "image/color"

// ModifiedBadlandsPlateau ...
type ModifiedBadlandsPlateau struct{}

// Temperature ...
func (ModifiedBadlandsPlateau) Temperature() float64 {
	return 2
}

// Rainfall ...
func (ModifiedBadlandsPlateau) Rainfall() float64 {
	return 0
}

// Depth ...
func (ModifiedBadlandsPlateau) Depth() float64 {
	return 0.45
}

// Scale ...
func (ModifiedBadlandsPlateau) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (ModifiedBadlandsPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x80, B: 0x9e, A: 0xa5}
}

// Tags ...
func (ModifiedBadlandsPlateau) Tags() []string {
	return []string{"animal", "mesa", "monster", "mutated", "overworld", "plateau", "stone", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (ModifiedBadlandsPlateau) String() string {
	return "mesa_plateau_mutated"
}

// EncodeBiome ...
func (ModifiedBadlandsPlateau) EncodeBiome() int {
	return 167
}
