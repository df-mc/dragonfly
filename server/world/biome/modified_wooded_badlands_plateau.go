package biome

import "image/color"

type ModifiedWoodedBadlandsPlateau struct{}

func (ModifiedWoodedBadlandsPlateau) Temperature() float64 {
	return 2
}

func (ModifiedWoodedBadlandsPlateau) Rainfall() float64 {
	return 0
}

func (ModifiedWoodedBadlandsPlateau) Depth() float64 {
	return 0.45
}

func (ModifiedWoodedBadlandsPlateau) Scale() float64 {
	return 0.3
}

func (ModifiedWoodedBadlandsPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x80, B: 0x9e, A: 0xa5}
}

func (ModifiedWoodedBadlandsPlateau) Tags() []string {
	return []string{"animal", "mesa", "monster", "mutated", "overworld", "plateau", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

func (ModifiedWoodedBadlandsPlateau) String() string {
	return "mesa_plateau_stone_mutated"
}

func (ModifiedWoodedBadlandsPlateau) EncodeBiome() int {
	return 166
}
