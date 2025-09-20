package biome

import "image/color"

type SavannaPlateau struct{}

func (SavannaPlateau) Temperature() float64 {
	return 1
}

func (SavannaPlateau) Rainfall() float64 {
	return 0
}

func (SavannaPlateau) Depth() float64 {
	return 1.5
}

func (SavannaPlateau) Scale() float64 {
	return 0.025
}

func (SavannaPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x25, G: 0x90, B: 0xa8, A: 0xa5}
}

func (SavannaPlateau) Tags() []string {
	return []string{"animal", "monster", "overworld", "plateau", "savanna", "spawns_savanna_mobs", "spawns_warm_variant_farm_animals"}
}

func (SavannaPlateau) String() string {
	return "savanna_plateau"
}

func (SavannaPlateau) EncodeBiome() int {
	return 36
}
