package biome

// ModifiedJungle ...
type ModifiedJungle struct{}

// Temperature ...
func (ModifiedJungle) Temperature() float64 {
	return 0.95
}

// Rainfall ...
func (ModifiedJungle) Rainfall() float64 {
	return 0.9
}

// Ash ...
func (ModifiedJungle) Ash() float64 {
	return 0
}

// WhiteAsh ...
func (ModifiedJungle) WhiteAsh() float64 {
	return 0
}

// BlueSpores ...
func (ModifiedJungle) BlueSpores() float64 {
	return 0
}

// RedSpores ...
func (ModifiedJungle) RedSpores() float64 {
	return 0
}

// String ...
func (ModifiedJungle) String() string {
	return "jungle_mutated"
}

// EncodeBiome ...
func (ModifiedJungle) EncodeBiome() int {
	return 149
}
