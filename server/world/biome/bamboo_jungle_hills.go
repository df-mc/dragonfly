package biome

import "image/color"

// BambooJungleHills ...
type BambooJungleHills struct{}

// Temperature ...
func (BambooJungleHills) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (BambooJungleHills) Rainfall() float64 {
	return 0.9
}

// Depth ...
func (BambooJungleHills) Depth() float64 {
	return 0.45
}

// Scale ...
func (BambooJungleHills) Scale() float64 {
	return 0.3
}

// WaterColour ...
func (BambooJungleHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1b, G: 0x9e, B: 0xd8, A: 0xa5}
}

// Tags ...
func (BambooJungleHills) Tags() []string {
	return []string{"animal", "bamboo", "hills", "jungle", "monster", "overworld", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

// String ...
func (BambooJungleHills) String() string {
	return "bamboo_jungle_hills"
}

// EncodeBiome ...
func (BambooJungleHills) EncodeBiome() int {
	return 49
}
