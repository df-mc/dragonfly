package biome

// CrimsonForest ...
type CrimsonForest struct{}

// Temperature ...
func (CrimsonForest) Temperature() float64 {
	return 2
}

// Rainfall ...
func (CrimsonForest) Rainfall() float64 {
	return 0
}

// String ...
func (CrimsonForest) String() string {
	return "Crimson Forest"
}

// EncodeBiome ...
func (CrimsonForest) EncodeBiome() int {
	return 179
}
