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

// Spores ...
func (CrimsonForest) Spores() (blueSpores float64, redSpores float64) {
	return 0, 0.25
}

// String ...
func (CrimsonForest) String() string {
	return "crimson_forest"
}

// EncodeBiome ...
func (CrimsonForest) EncodeBiome() int {
	return 179
}
