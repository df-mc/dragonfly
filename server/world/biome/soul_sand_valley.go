package biome

import "image/color"

// SoulSandValley ...
type SoulSandValley struct{}

// Temperature ...
func (SoulSandValley) Temperature() float64 {
	return 2
}

// Rainfall ...
func (SoulSandValley) Rainfall() float64 {
	return 0
}

// Depth ...
func (SoulSandValley) Depth() float64 {
	return 0.1
}

// Scale ...
func (SoulSandValley) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (SoulSandValley) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

// Tags ...
func (SoulSandValley) Tags() []string {
	return []string{"nether", "soulsand_valley", "spawn_ghast", "spawn_endermen", "spawns_warm_variant_farm_animals"}
}

// String ...
func (SoulSandValley) String() string {
	return "soulsand_valley"
}

// EncodeBiome ...
func (SoulSandValley) EncodeBiome() int {
	return 178
}
