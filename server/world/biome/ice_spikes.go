package biome

import "image/color"

// IceSpikes ...
type IceSpikes struct{}

// Temperature ...
func (IceSpikes) Temperature() float64 {
	return 0
}

// Rainfall ...
func (IceSpikes) Rainfall() float64 {
	return 1
}

// Depth ...
func (IceSpikes) Depth() float64 {
	return 0.425
}

// Scale ...
func (IceSpikes) Scale() float64 {
	return 0.45
}

// WaterColour ...
func (IceSpikes) WaterColour() color.RGBA {
	return color.RGBA{R: 0x60, G: 0xb7, B: 0xff, A: 0xa6}
}

// Tags ...
func (IceSpikes) Tags() []string {
	return []string{"frozen", "ice_plains", "monster", "mutated", "overworld", "spawns_cold_variant_farm_animals", "spawns_cold_variant_frogs", "spawns_white_rabbits"}
}

// String ...
func (IceSpikes) String() string {
	return "ice_plains_spikes"
}

// EncodeBiome ...
func (IceSpikes) EncodeBiome() int {
	return 140
}
