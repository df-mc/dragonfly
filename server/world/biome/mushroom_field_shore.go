package biome

import "image/color"

type MushroomFieldShore struct{}

func (MushroomFieldShore) Temperature() float64 {
	return 0.9
}

func (MushroomFieldShore) Rainfall() float64 {
	return 1
}

func (MushroomFieldShore) Depth() float64 {
	return 0
}

func (MushroomFieldShore) Scale() float64 {
	return 0.025
}

func (MushroomFieldShore) WaterColour() color.RGBA {
	return color.RGBA{R: 0x81, G: 0x81, B: 0x93, A: 0xa5}
}

func (MushroomFieldShore) Tags() []string {
	return []string{"mooshroom_island", "overworld", "shore", "spawns_without_patrols"}
}

func (MushroomFieldShore) String() string {
	return "mushroom_island_shore"
}

func (MushroomFieldShore) EncodeBiome() int {
	return 15
}
