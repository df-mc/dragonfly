package biome

// SavannaPlateau ...
type SavannaPlateau struct{}

// Temperature ...
func (SavannaPlateau) Temperature() float64 {
	return 1
}

// Rainfall ...
func (SavannaPlateau) Rainfall() float64 {
	return 0
}

// String ...
func (SavannaPlateau) String() string {
	return "Savanna Plateau"
}

// EncodeBiome ...
func (SavannaPlateau) EncodeBiome() int {
	return 36
}
