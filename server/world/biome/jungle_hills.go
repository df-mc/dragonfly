package biome

import "image/color"

// JungleHills ...
type JungleHills struct{}

// Temperature ...
func (JungleHills) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (JungleHills) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (JungleHills) Depth() float64 {
	return 0.45
}

// Scale ...
func (JungleHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (JungleHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1b, G: 0x9e, B: 0xd8, A: 0xa5}
}

// Tags ...
func (JungleHills) Tags() []string {
	return []string{"animal", "hills", "jungle", "monster", "overworld", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (JungleHills) String() string {
	return "jungle_hills"
}

// EncodeBiome ...
func (JungleHills) EncodeBiome() int {
	return 22
}
