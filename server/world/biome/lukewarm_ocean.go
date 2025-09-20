package biome

import "image/color"

type LukewarmOcean struct{}

func (LukewarmOcean) Temperature() float64 {
	return 0.5
}

func (LukewarmOcean) Rainfall() float64 {
	return 0.5
}

func (LukewarmOcean) Depth() float64 {
	return -1
}

func (LukewarmOcean) Scale() float64 {
	return 0.1
}

func (LukewarmOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x96, B: 0xdb, A: 0xa5}
}

func (LukewarmOcean) Tags() []string {
	return []string{"lukewarm", "monster", "ocean", "overworld", "spawns_warm_variant_farm_animals"}
}

func (LukewarmOcean) String() string {
	return "lukewarm_ocean"
}

func (LukewarmOcean) EncodeBiome() int {
	return 42
}
