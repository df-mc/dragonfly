package biome

import "image/color"

// ShatteredSavannaPlateau ...
type ShatteredSavannaPlateau struct{}

// Temperature ...
func (ShatteredSavannaPlateau) Temperature() float64 {
	return 1
}

// Rainfall ...
func (ShatteredSavannaPlateau) Rainfall() float64 {
	return 0.5
}

// Depth ...
func (ShatteredSavannaPlateau) Depth() float64 {
	return 1.05
}

// Scale ...
func (ShatteredSavannaPlateau) Scale() float64 {
	return 1.212
}

// WaterColour ...
func (ShatteredSavannaPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (ShatteredSavannaPlateau) Tags() []string {
	return []string{"animal", "monster", "mutated", "overworld", "plateau", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (ShatteredSavannaPlateau) String() string {
	return "savanna_plateau_mutated"
}

// EncodeBiome ...
func (ShatteredSavannaPlateau) EncodeBiome() int {
	return 164
}
