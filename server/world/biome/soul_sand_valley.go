package biome

import "image/color"

type SoulSandValley struct{}

func (SoulSandValley) Temperature() float64 {
	return 2
}

func (SoulSandValley) Rainfall() float64 {
	return 0
}

func (SoulSandValley) Ash() (ash float64, whiteAsh float64) {
	return 0.05, 0
}

func (SoulSandValley) Depth() float64 {
	return 0.1
}

func (SoulSandValley) Scale() float64 {
	return 0.2
}

func (SoulSandValley) WaterColour() color.RGBA {
	return color.RGBA{R: 0x90, G: 0x59, B: 0x57, A: 0xa5}
}

func (SoulSandValley) Tags() []string {
	return []string{"nether", "soulsand_valley", "spawn_ghast", "spawn_endermen", "spawns_warm_variant_farm_animals"}
}

func (SoulSandValley) String() string {
	return "soulsand_valley"
}

func (SoulSandValley) EncodeBiome() int {
	return 178
}
