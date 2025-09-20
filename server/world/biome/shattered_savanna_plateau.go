package biome

import "image/color"

type ShatteredSavannaPlateau struct{}

func (ShatteredSavannaPlateau) Temperature() float64 {
	return 1
}

func (ShatteredSavannaPlateau) Rainfall() float64 {
	return 0.5
}

func (ShatteredSavannaPlateau) Depth() float64 {
	return 1.05
}

func (ShatteredSavannaPlateau) Scale() float64 {
	return 1.212
}

func (ShatteredSavannaPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

func (ShatteredSavannaPlateau) Tags() []string {
	return []string{"animal", "monster", "mutated", "overworld", "plateau", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

func (ShatteredSavannaPlateau) String() string {
	return "savanna_plateau_mutated"
}

func (ShatteredSavannaPlateau) EncodeBiome() int {
	return 164
}
