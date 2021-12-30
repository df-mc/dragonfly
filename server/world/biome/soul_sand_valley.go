package biome

// SoulSandValley ...
type SoulSandValley struct{}

// Temperature ...
func (SoulSandValley) Temperature() float64 {
	return 2
}

// Rainfall ...
func (SoulSandValley) Rainfall() float64 {
	return 0
}

// Ash ...
func (SoulSandValley) Ash() (ash float64, whiteAsh float64) {
	return 0.05, 0
}

// String ...
func (SoulSandValley) String() string {
	return "soulsand_valley"
}

// EncodeBiome ...
func (SoulSandValley) EncodeBiome() int {
	return 178
}
