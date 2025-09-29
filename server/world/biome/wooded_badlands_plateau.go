package biome

import "image/color"

type WoodedBadlandsPlateau struct{}

func (WoodedBadlandsPlateau) Temperature() float64 {
	return 2
}

func (WoodedBadlandsPlateau) Rainfall() float64 {
	return 0
}

func (WoodedBadlandsPlateau) Depth() float64 {
	return 1.5
}

func (WoodedBadlandsPlateau) Scale() float64 {
	return 0.025
}

func (WoodedBadlandsPlateau) WaterColour() color.RGBA {
	return color.RGBA{R: 0x55, G: 0x80, B: 0x9e, A: 0xa5}
}

func (WoodedBadlandsPlateau) Tags() []string {
	return []string{"animal", "mesa", "monster", "overworld", "plateau", "rare", "stone", "spawns_mesa_mobs", "spawns_warm_variant_farm_animals"}
}

func (WoodedBadlandsPlateau) String() string {
	return "mesa_plateau_stone"
}

func (WoodedBadlandsPlateau) EncodeBiome() int {
	return 38
}
