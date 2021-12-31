package biome

// ShatteredSavannaPlateau ...
type ShatteredSavannaPlateau struct{}

// Temperature ...
func (ShatteredSavannaPlateau) Temperature() float64 {
	return 1
}

// Rainfall ...
func (ShatteredSavannaPlateau) Rainfall() float64 {
	return 0.5
}

// String ...
func (ShatteredSavannaPlateau) String() string {
	return "savanna_plateau_mutated"
}

// EncodeBiome ...
func (ShatteredSavannaPlateau) EncodeBiome() int {
	return 164
}
