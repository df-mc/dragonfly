package biome

import "image/color"

// NetherWastes ...
type NetherWastes struct{}

// Temperature ...
func (NetherWastes) Temperature() float64 {
	return 2
}

// Rainfall ...
func (NetherWastes) Rainfall() float64 {
	return 0
}

// Depth ...
func (NetherWastes) Depth() float64 {
	return 0.1
}

// Scale ...
func (NetherWastes) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (NetherWastes) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

// Tags ...
func (NetherWastes) Tags() []string {
	return []string{"nether", "nether_wastes", "spawn_endermen", "spawn_few_piglins", "spawn_ghast", "spawn_magma_cubes", "spawns_nether_mobs", "spawn_zombified_piglin", "spawns_warm_variant_farm_animals"}
}

// String ...
func (NetherWastes) String() string {
	return "hell"
}

// EncodeBiome ...
func (NetherWastes) EncodeBiome() int {
	return 8
}
