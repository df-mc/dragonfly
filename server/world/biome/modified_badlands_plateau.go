package biome

import "image/color"

type ModifiedBadlandsPlateau struct{}

func (ModifiedBadlandsPlateau) Temperature() float64 {
	return 2
}

func (ModifiedBadlandsPlateau) Rainfall() float64 {
	return 0
}

func (ModifiedBadlandsPlateau) Depth() float64 {
	return 0.45
}

func (ModifiedBadlandsPlateau) Scale() float64 {
	return 0.3
}

func (ModifiedBadlandsPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x80, B: 0x9e, A: 0xa5}
}

func (ModifiedBadlandsPlateau) Tags() []string {
	return []string{"animal", "mesa", "monster", "mutated", "overworld", "plateau", "stone", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

func (ModifiedBadlandsPlateau) String() string {
	return "mesa_plateau_mutated"
}

func (ModifiedBadlandsPlateau) EncodeBiome() int {
	return 167
}
