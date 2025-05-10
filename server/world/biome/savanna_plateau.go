package biome

import "image/color"

// SavannaPlateau ...
type SavannaPlateau struct{}

// Temperature ...
func (SavannaPlateau) Temperature() float64 {
	return 1
}

// Rainfall ...
func (SavannaPlateau) Rainfall() float64 {
	return 0
}

// Depth ...
func (SavannaPlateau) Depth() float64 {
	return 1.5
}

// Scale ...
func (SavannaPlateau) Scale() float64 {
	return 0.025
}

// WaterColour ...
func (SavannaPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x90, B: 0xa8, A: 0xa5}
}

// Tags ...
func (SavannaPlateau) Tags() []string {
	return []string{"animal", "monster", "overworld", "plateau", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (SavannaPlateau) String() string {
	return "savanna_plateau"
}

// EncodeBiome ...
func (SavannaPlateau) EncodeBiome() int {
	return 36
}
