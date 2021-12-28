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
func (SoulSandValley) Ash() float64 {
	return 0.05
}

// WhiteAsh ...
func (SoulSandValley) WhiteAsh() float64 {
	return 0
}

// String ...
func (SoulSandValley) String() string {
	return "soulsand_valley"
}

// EncodeBiome ...
func (SoulSandValley) EncodeBiome() int {
	return 178
}
