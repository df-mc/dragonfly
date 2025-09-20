package biome

import "image/color"

type MushroomFields struct{}

func (MushroomFields) Temperature() float64 {
	return 0.9
}

func (MushroomFields) Rainfall() float64 {
	return 1
}

func (MushroomFields) Depth() float64 {
	return 0.2
}

func (MushroomFields) Scale() float64 {
	return 0.3
}

func (MushroomFields) WaterColour() color.RGBA {
	return color.RGBA{R: 0x8a, G: 0x89, B: 0x97, A: 0xa5}
}

func (MushroomFields) Tags() []string {
	return []string{"mooshroom_island", "overworld", "spawns_without_patrols"}
}

func (MushroomFields) String() string {
	return "mushroom_island"
}

func (MushroomFields) EncodeBiome() int {
	return 14
}
