package biome

import "image/color"

// SnowyTaiga ...
type SnowyTaiga struct{}

// Temperature ...
func (SnowyTaiga) Temperature() float64 {
	return -0.5
}

// Rainfall ...
func (SnowyTaiga) Rainfall() float64 {
	return 0.4
}

// Depth ...
func (SnowyTaiga) Depth() float64 {
	return 0.2
}

// Scale ...
func (SnowyTaiga) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (SnowyTaiga) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x5e, B: 0x83, A: 0xa5}
}

// Tags ...
func (SnowyTaiga) Tags() []string {
	return []string{"animal", "cold", "forest", "monster", "overworld", "taiga", "has_structure_trail_ruins", "spawns_cold_variant_farm_animals"}
}

// String ...
func (SnowyTaiga) String() string {
	return "cold_taiga"
}

// EncodeBiome ...
func (SnowyTaiga) EncodeBiome() int {
	return 30
}
