package biome

import "image/color"

type DeepLukewarmOcean struct{}

func (DeepLukewarmOcean) Temperature() float64 {
	return 0.5
}

func (DeepLukewarmOcean) Rainfall() float64 {
	return 0.5
}

func (DeepLukewarmOcean) Depth() float64 {
	return -1.8
}

func (DeepLukewarmOcean) Scale() float64 {
	return 0.1
}

func (DeepLukewarmOcean) WaterColour() color.RGBA {
	return color.RGBA{R: 0x0d, G: 0x96, B: 0xdb, A: 0xa5}
}

func (DeepLukewarmOcean) Tags() []string {
	return []string{"deep", "lukewarm", "monster", "ocean", "overworld", "spawns_warm_variant_farm_animals"}
}

func (DeepLukewarmOcean) String() string {
	return "deep_lukewarm_ocean"
}

func (DeepLukewarmOcean) EncodeBiome() int {
	return 43
}
