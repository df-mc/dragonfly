package biome

import "image/color"

type River struct{}

func (River) Temperature() float64 {
	return 0.5
}

func (River) Rainfall() float64 {
	return 0.5
}

func (River) Depth() float64 {
	return -0.5
}

func (River) Scale() float64 {
	return 0
}

func (River) WaterColour() color.RGBA {
	return color.RGBA{R: 0x00, G: 0x84, B: 0xff, A: 0xa5}
}

func (River) Tags() []string {
	return []string{"overworld", "spawns_more_frequent_drowned", "spawns_reduced_water_ambient_mobs", "spawns_river_mobs", "river"}
}

func (River) String() string {
	return "river"
}

func (River) EncodeBiome() int {
	return 7
}
