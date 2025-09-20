package biome

import "image/color"

type MangroveSwamp struct{}

func (MangroveSwamp) Temperature() float64 {
	return 0.8
}

func (MangroveSwamp) Rainfall() float64 {
	return 0.9
}

func (MangroveSwamp) Depth() float64 {
	return -0.2
}

func (MangroveSwamp) Scale() float64 {
	return 0.1
}

func (MangroveSwamp) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (MangroveSwamp) Tags() []string {
	return []string{"mangrove_swamp", "overworld", "monster", "spawns_slimes_on_surface", "spawns_warm_variant_farm_animals", "spawns_warm_variant_frogs"}
}

func (MangroveSwamp) String() string {
	return "mangrove_swamp"
}

func (MangroveSwamp) EncodeBiome() int {
	return 191
}
