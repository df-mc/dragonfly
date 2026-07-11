package biome

import "image/color"

// OldGrowthPineTaiga ...
type OldGrowthPineTaiga struct{}

// Temperature ...
func (OldGrowthPineTaiga) Temperature() float64 {
	return 0.3
}

// Rainfall ...
func (OldGrowthPineTaiga) Rainfall() float64 {
	return 0.8
}

// Depth ...
func (OldGrowthPineTaiga) Depth() float64 {
	return 0.2
}

// Scale ...
func (OldGrowthPineTaiga) Scale() float64 {
	return 0.2
}

// WaterColour ...
func (OldGrowthPineTaiga) WaterColour() color.RGBA {
	return color.RGBA{R: 0x2d, G: 0x6d, B: 0x77, A: 0xa5}
}

// Tags ...
func (OldGrowthPineTaiga) Tags() []string {
	return []string{"animal", "forest", "mega", "monster", "overworld", "rare", "taiga", "has_structure_trail_ruins", "spawns_cold_variant_farm_animals"}
}

// String ...
func (OldGrowthPineTaiga) String() string {
	return "mega_taiga"
}

// EncodeBiome ...
func (OldGrowthPineTaiga) EncodeBiome() int {
	return 32
}
