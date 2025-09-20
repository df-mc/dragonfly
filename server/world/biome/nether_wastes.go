package biome

import "image/color"

type NetherWastes struct{}

func (NetherWastes) Temperature() float64 {
	return 2
}

func (NetherWastes) Rainfall() float64 {
	return 0
}

func (NetherWastes) Depth() float64 {
	return 0.1
}

func (NetherWastes) Scale() float64 {
	return 0.2
}

func (NetherWastes) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

func (NetherWastes) Tags() []string {
	return []string{"nether", "nether_wastes", "spawn_endermen", "spawn_few_piglins", "spawn_ghast", "spawn_magma_cubes", "spawns_nether_mobs", "spawn_zombified_piglin", "spawns_warm_variant_farm_animals"}
}

func (NetherWastes) String() string {
	return "hell"
}

func (NetherWastes) EncodeBiome() int {
	return 8
}
