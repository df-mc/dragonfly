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

// String ...
func (ModifiedJungle) String() string {
	return "jungle_mutated"
}

// EncodeBiome ...
func (ModifiedJungle) EncodeBiome() int {
	return 149
}
