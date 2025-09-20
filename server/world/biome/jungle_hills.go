package biome

import "image/color"

type JungleHills struct{}

func (JungleHills) Temperature() float64 {
	return 0.95
}

func (JungleHills) Rainfall() float64 {
	return 0.9
}

func (JungleHills) Depth() float64 {
	return 0.45
}

func (JungleHills) Scale() float64 {
	return 0.3
}

func (JungleHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1b, G: 0x9e, B: 0xd8, A: 0xa5}
}

func (JungleHills) Tags() []string {
	return []string{"animal", "hills", "jungle", "monster", "overworld", "spawns_jungle_mobs", "spawns_warm_variant_farm_animals"}
}

func (JungleHills) String() string {
	return "jungle_hills"
}

func (JungleHills) EncodeBiome() int {
	return 22
}
