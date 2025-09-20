package biome

import "image/color"

type SnowyTaiga struct{}

func (SnowyTaiga) Temperature() float64 {
	return -0.5
}

func (SnowyTaiga) Rainfall() float64 {
	return 0.4
}

func (SnowyTaiga) Depth() float64 {
	return 0.2
}

func (SnowyTaiga) Scale() float64 {
	return 0.2
}

func (SnowyTaiga) WaterColour() color.RGBA {
	return color.RGBA{R: 0x20, G: 0x5e, B: 0x83, A: 0xa5}
}

func (SnowyTaiga) Tags() []string {
	return []string{"animal", "cold", "forest", "monster", "overworld", "taiga", "has_structure_trail_ruins", "spawns_cold_variant_farm_animals"}
}

func (SnowyTaiga) String() string {
	return "cold_taiga"
}

func (SnowyTaiga) EncodeBiome() int {
	return 30
}
