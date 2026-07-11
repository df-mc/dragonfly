package biome

import "image/color"

// BasaltDeltas ...
type BasaltDeltas struct{}

// Temperature ...
func (BasaltDeltas) Temperature() float64 {
	return 2
}

// Rainfall ...
func (BasaltDeltas) Rainfall() float64 {
	return 0
}

// Depth ...
func (BasaltDeltas) Depth() float64 {
	return 0.1
}

// Scale ...
func (BasaltDeltas) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (BasaltDeltas) WaterColour() color.RGBA {
	return color.RGBA{R: 0x3f, G: 0x76, B: 0xe4, A: 0xa5}
}

// Tags ...
func (BasaltDeltas) Tags() []string {
	return []string{"nether", "basalt_deltas", "spawn_many_magma_cubes", "spawn_ghast", "spawns_warm_variant_farm_animals"}
}

// String ...
func (BasaltDeltas) String() string {
	return "basalt_deltas"
}

// EncodeBiome ...
func (BasaltDeltas) EncodeBiome() int {
	return 181
}
