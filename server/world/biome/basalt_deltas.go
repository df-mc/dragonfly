package biome

import "image/color"

type BasaltDeltas struct{}

func (BasaltDeltas) Temperature() float64 {
	return 2
}

func (BasaltDeltas) Rainfall() float64 {
	return 0
}

func (BasaltDeltas) Ash() (ash float64, whiteAsh float64) {
	return 0, 2
}

func (BasaltDeltas) Depth() float64 {
	return 0.1
}

func (BasaltDeltas) Scale() float64 {
	return 0.2
}

func (BasaltDeltas) WaterColour() color.RGBA {
	return color.RGBA{R: 0x3f, G: 0x76, B: 0xe4, A: 0xa5}
}

func (BasaltDeltas) Tags() []string {
	return []string{"nether", "basalt_deltas", "spawn_many_magma_cubes", "spawn_ghast", "spawns_warm_variant_farm_animals"}
}

func (BasaltDeltas) String() string {
	return "basalt_deltas"
}

func (BasaltDeltas) EncodeBiome() int {
	return 181
}
