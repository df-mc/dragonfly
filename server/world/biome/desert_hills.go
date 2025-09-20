package biome

import "image/color"

type DesertHills struct{}

func (DesertHills) Temperature() float64 {
	return 2
}

func (DesertHills) Rainfall() float64 {
	return 0
}

func (DesertHills) Depth() float64 {
	return 0.45
}

func (DesertHills) Scale() float64 {
	return 0.3
}

func (DesertHills) WaterColour() color.RGBA {
	return color.RGBA{R: 0x1a, G: 0x7a, B: 0xa1, A: 0xa5}
}

func (DesertHills) Tags() []string {
	return []string{"desert", "hills", "monster", "overworld", "spawns_gold_rabbits", "spawns_warm_variant_farm_animals"}
}

func (DesertHills) String() string {
	return "desert_hills"
}

func (DesertHills) EncodeBiome() int {
	return 17
}
